#!/bin/bash

function uninstall() {
  nodeadm uninstall -d

  iptables -F && iptables -t nat -F && iptables -t mangle -F && iptables -X

  set +e
  rm -rf /etc/eks/ /etc/kubernetes
  rm -rf /etc/cni/net.d /opt/cni
  rm -rf /etc/systemd/system/kubelet.service
  rm -rf /usr/bin/containerd /usr/bin/kubelet /usr/local/bin/aws* /usr/local/bin/kubectl
  rm -rf /var/log/containers
  rm -rf /var/log/nodeadm-*.log
  set -e
}