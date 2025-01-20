#!/bin/bash

exec > >(tee -ia /var/log/nodeadm-upgrade.log)
exec 2> >(tee -ia /var/log/nodeadm-upgrade.log >& 2)
exec 22>> /var/log/nodeadm-upgrade.log

export BASH_XTRACEFD="22"
set -ex

kubernetes_version=$1
config_file=$2

root_path=$3
proxy_configured=$4
proxy_http=$5
proxy_https=$6
proxy_no=$7

export PATH="$PATH:$root_path/bin"

if [ -n "$proxy_no" ]; then
  export NO_PROXY=$proxy_no
  export no_proxy=$proxy_no
fi

if [ -n "$proxy_http" ]; then
  export HTTP_PROXY=$proxy_http
  export http_proxy=$proxy_http
fi

if [ -n "$proxy_https" ]; then
  export https_proxy=$proxy_https
  export HTTPS_PROXY=$proxy_https
fi

export KUBECONFIG=/var/lib/kubelet/kubeconfig

current_node_name=$(cat /etc/hostname)

run_upgrade() {
    echo "running upgrade process on $current_node_name"

    if ! [ -f "$root_path/sentinel_kubernetes_version" ]; then
      echo "$kubernetes_version" > "$root_path/sentinel_kubernetes_version"
      echo "upgrade is a no-op; created sentinel file with version: $kubernetes_version"
      return
    fi

    old_version=$(cat "$root_path/sentinel_kubernetes_version")
    echo "found last deployed version $old_version"

    current_version=$kubernetes_version
    echo "found current deployed version $current_version"

    # Check if the current nodeadm version is equal to the stored nodeadm version.
    # If yes, do nothing.
    if [ "$current_version" = "$old_version" ]
    then
      echo "node is on latest version: $current_version"
      return
    fi

    # Backup admin.conf if it exists, as nodeadm upgrade wipes /etc/kubernetes
    if [ -f /etc/kubernetes/admin.conf ]; then
      cp /etc/kubernetes/admin.conf /opt/nodeadmutil/admin.conf.bak
      echo "backed up /etc/kubernetes/admin.conf to /opt/nodeadmutil/admin.conf.bak"
    fi

    # Upgrade loop, runs until stored and current match
    until [ "$current_version" = "$old_version" ]
    do
        upgrade_command="nodeadm upgrade -c file://$config_file -d $kubernetes_version --skip pod-validation"

        if [ "$proxy_configured" = true ]; then
          up=("nodeadm upgrade -c file://$config_file -d $kubernetes_version --skip pod-validation")
          upgrade_command="${up[*]}"
        fi

        echo "upgrading node from $old_version to $current_version using command: $upgrade_command"

        # Update current kubernetes version
        if [ "$proxy_configured" = true ]; then
          if sudo -E bash -c "$upgrade_command"
          then
              echo "$current_version" > "$root_path/sentinel_kubernetes_version"
              old_version=$current_version
              echo "upgrade success"
          else
              echo "upgrade failed, retrying in 60 seconds"
              sleep 60
          fi
        else
          if $upgrade_command
          then
              echo "$current_version" > "$root_path/sentinel_kubernetes_version"
              old_version=$current_version
              echo "upgrade success"
          else
              echo "upgrade failed, retrying in 60 seconds"
              sleep 60
          fi
        fi
    done
}

run_upgrade

echo "Re-initializing node"
bash "$root_path"/scripts/nodeadm-init.sh "$config_file" "$root_path" "$proxy_configured" "$proxy_http" "$proxy_https" "$proxy_no"

# Restore admin.conf if it was backed up
if [ -f /opt/nodeadmutil/admin.conf.bak ]; then
  mv /opt/nodeadmutil/admin.conf.bak /etc/kubernetes/admin.conf
  echo "restored /etc/kubernetes/admin.conf from /opt/nodeadmutil/admin.conf.bak"
fi
