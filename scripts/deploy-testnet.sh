#!/usr/bin/env bash

if [[ ! -d ../keys ]];then
  mkdir ../keys
fi
./1_clean_nodes.sh
./4_rebuild_nodes_testnet.sh




