#!/bin/bash

exec > >(tee -ia /var/log/nodeadm-reset.log)
exec 2> >(tee -ia /var/log/nodeadm-reset.log >& 2)
exec 21>> /var/log/nodeadm-reset.log

export BASH_XTRACEFD="21"
set -ex

root_path=$1

source "$root_path"/scripts/uninstall.sh

export PATH="$PATH:$root_path/bin"

uninstall $root_path