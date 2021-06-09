#!/usr/bin/env bash

if [[ ! -d ../keys ]];then
  mkdir ../keys
fi
./1_clean_nodes.sh
./2_redeploy_contracts.sh && \
./3_edit_configs.sh && \
./4_rebuild_nodes.sh




