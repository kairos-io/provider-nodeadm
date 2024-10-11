#!/bin/bash

exec > >(tee -ia /var/log/nodeadm-reset.log)
exec 2> >(tee -ia /var/log/nodeadm-reset.log >& 2)
exec 19>> /var/log/nodeadm-reset.log

export BASH_XTRACEFD="19"
set -ex

source ./uninstall.sh

root_path=$1

export PATH="$PATH:$root_path/bin"

uninstall