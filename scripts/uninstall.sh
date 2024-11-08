#!/bin/bash

function uninstall() {
  set +e

  root_path=$1

  systemctl stop kubelet
  systemctl stop containerd

  pkill -9 kubelet || true
  pkill -9 containerd || true

  umount -l /var/lib/kubelet
  umount -l /var/lib/containerd

  rm -rf /var/lib/kubelet && rm -rf $root_path/var/lib/kubelet
  rm -rf /var/lib/containerd && rm -rf $root_path/var/lib/containerd

  nodeadm uninstall -d

  iptables -F && iptables -t nat -F && iptables -t mangle -F && iptables -X

  rm -rf /var/log/containers
  rm -rf /var/log/nodeadm-*.log

  set -e
}