#!/usr/bin/env bash

if [[ ! -d ../keys ]];then
  mkdir ../keys
fi
./1_clean_nodes.sh
./2_redeploy_contracts.sh && \
./3_build_bsn.sh && \
./4_build_config.sh && \
./5_rebuild_nodes.sh




