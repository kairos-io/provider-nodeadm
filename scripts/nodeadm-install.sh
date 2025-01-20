#!/bin/bash

exec > >(tee -ia /var/log/nodeadm-install.log)
exec 2> >(tee -ia /var/log/nodeadm-install.log >& 2)
exec 20>> /var/log/nodeadm-install.log

export BASH_XTRACEFD="20"
set -ex

kubernetes_version=$1
credential_provider=$2

root_path=$3
proxy_configured=$4
proxy_http=$5
proxy_https=$6
proxy_no=$7

source "$root_path"/scripts/uninstall.sh

export PATH="$PATH:$root_path/bin"

function uninstall_and_retry() {
  echo "nodeadm install failed, applying reset";
  set +e
  uninstall $root_path
  set -e
  echo "retrying in 10s"
  sleep 10;
}

if [ "$proxy_configured" = true ]; then
  until HTTP_PROXY=$proxy_http http_proxy=$proxy_http HTTPS_PROXY=$proxy_https https_proxy=$proxy_https NO_PROXY=$proxy_no no_proxy=$proxy_no nodeadm install $kubernetes_version -p $credential_provider -d > /dev/null
  do
    uninstall_and_retry
  done;
else
  until nodeadm install $kubernetes_version -p $credential_provider -d > /dev/null
  do
    uninstall_and_retry
  done;
fi