#!/bin/bash

exec > >(tee -ia /var/log/nodeadm-install.log)
exec 2> >(tee -ia /var/log/nodeadm-install.log >& 2)
exec 19>> /var/log/nodeadm-install.log

export BASH_XTRACEFD="19"
set -ex

KUBERNETES_VERSION=$1
CREDENTIAL_PROVIDER=$2

PROXY_CONFIGURED=$3
proxy_http=$4
proxy_https=$5
proxy_no=$6

function uninstall_and_retry() {
  echo "nodeadm install failed, applying reset";
  uninstall
  echo "retrying in 10s"
  sleep 10;
}

function uninstall() {
  nodeadm uninstall -d
  iptables -F && iptables -t nat -F && iptables -t mangle -F && iptables -X
  rm -rf /etc/kubernetes
  rm -rf /etc/cni/net.d /opt/cni
  rm -rf /usr/bin/containerd /usr/bin/kubelet /etc/systemd/system/kubelet.service
  rm -rf /usr/local/bin/aws* /usr/local/bin/kubectl /etc/eks/
}

if [ "$PROXY_CONFIGURED" = true ]; then
  until HTTP_PROXY=$proxy_http http_proxy=$proxy_http HTTPS_PROXY=$proxy_https https_proxy=$proxy_https NO_PROXY=$proxy_no no_proxy=$proxy_no nodeadm install $KUBERNETES_VERSION -p $CREDENTIAL_PROVIDER -d > /dev/null
  do
    uninstall_and_retry
  done;
else
  until nodeadm install $KUBERNETES_VERSION -p $CREDENTIAL_PROVIDER -d > /dev/null
  do
    uninstall_and_retry
  done;
fi