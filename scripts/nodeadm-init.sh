#!/bin/bash

exec > >(tee -ia /var/log/nodeadm-init.log)
exec 2> >(tee -ia /var/log/nodeadm-init.log >& 2)
exec 19>> /var/log/nodeadm-init.log

export BASH_XTRACEFD="19"
set -ex

config_file=$1

root_path=$2
proxy_configured=$3
proxy_http=$4
proxy_https=$5
proxy_no=$6

export PATH="$PATH:$root_path/bin"

function retry_init() {
  echo "nodeadm init failed, retrying in 10s";
  sleep 10;
}

if [ "$proxy_configured" = true ]; then
  until HTTP_PROXY=$proxy_http http_proxy=$proxy_http HTTPS_PROXY=$proxy_https https_proxy=$proxy_https NO_PROXY=$proxy_no no_proxy=$proxy_no nodeadm init -c file://$config_file -d > /dev/null
  do
    retry_init
  done;
else
  until nodeadm init -c file://$config_file -d > /dev/null
  do
    retry_init
  done;
fi