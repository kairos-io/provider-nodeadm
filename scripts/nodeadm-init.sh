#!/bin/bash

exec > >(tee -ia /var/log/nodeadm-init.log)
exec 2> >(tee -ia /var/log/nodeadm-init.log >& 2)
exec 19>> /var/log/nodeadm-init.log

export BASH_XTRACEFD="19"
set -ex

CONFIG_FILE=$1

PROXY_CONFIGURED=$2
proxy_http=$3
proxy_https=$4
proxy_no=$5

function retry_init() {
  echo "nodeadm init failed, retrying in 10s";
  sleep 10;
}

if [ "$PROXY_CONFIGURED" = true ]; then
  until HTTP_PROXY=$proxy_http http_proxy=$proxy_http HTTPS_PROXY=$proxy_https https_proxy=$proxy_https NO_PROXY=$proxy_no no_proxy=$proxy_no nodeadm init -c file://$CONFIG_FILE -d > /dev/null
  do
    retry_init
  done;
else
  until /opt/nodeadm/bin/nodeadm init -c file://$CONFIG_FILE -d > /dev/null
  do
    retry_init
  done;
fi