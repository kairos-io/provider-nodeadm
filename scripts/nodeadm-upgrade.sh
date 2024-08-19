#!/bin/bash

exec > >(tee -ia /var/log/nodeadm-upgrade.log)
exec 2> >(tee -ia /var/log/nodeadm-upgrade.log >& 2)
exec 19>> /var/log/nodeadm-upgrade.log

set -x

KUBERNETES_VERSION=$1

PROXY_CONFIGURED=$2
proxy_http=$3
proxy_https=$4
proxy_no=$5

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

CURRENT_NODE_NAME=$(cat /etc/hostname)

run_upgrade() {
    echo "running upgrade process on $CURRENT_NODE_NAME"

    old_version=$(cat /opt/nodeadm/sentinel_kubernetes_version)
    echo "found last deployed version $old_version"

    current_version=$KUBERNETES_VERSION
    echo "found current deployed version $current_version"

    # Check if the current nodeadm version is equal to the stored nodeadm version.
    # If yes, do nothing.
    if [ "$current_version" = "$old_version" ]
    then
      echo "node is on latest version: $current_version"
      exit 0
    fi

    # Upgrade loop, runs until stored and current match
    until [ "$current_version" = "$old_version" ]
    do
        up=("/opt/nodeadm/bin/nodeadm upgrade $KUBERNETES_VERSION -d")
        upgrade_command="${up[*]}"

        echo "upgrading node from $old_version to $current_version using command: $upgrade_command"

        if sudo -E bash -c "$upgrade_command"
        then
            # Update current kubernetes version
            echo "$current_version" > /opt/nodeadm/sentinel_kubernetes_version
            old_version=$current_version

            echo "upgrade success"
        else
            echo "upgrade failed, retrying in 60 seconds"
            sleep 60
        fi
    done
}

run_upgrade