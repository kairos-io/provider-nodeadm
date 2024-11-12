#!/bin/bash

function uninstall() {
  root_path=$1

  systemctl stop kubelet
  systemctl stop containerd

  # Find all mount points under /var/lib/kubelet/pods and unmount them
  find /var/lib/kubelet/pods -type d -exec bash -c 'mountpoint -q "$1" && umount -l "$1"' _ {} \;

  rm -rf /var/lib/kubelet && rm -rf $root_path/var/lib/kubelet
  rm -rf /var/lib/containerd && rm -rf $root_path/var/lib/containerd

  nodeadm uninstall -d

  iptables -F && iptables -t nat -F && iptables -t mangle -F && iptables -X

  rm -rf /var/log/containers
  rm -rf /var/log/nodeadm-*.log
}