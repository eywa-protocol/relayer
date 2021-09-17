#!/usr/bin/env bash

function usage()
  {
    echo "Usage: $0 local|testnet"
    exit 1
  }

type=${1}

if  [ "$type" != "local" ] && [ "$type" != "testnet" ]; then
    usage
else

    ./1_clean_nodes.sh && \
    ./2_redeploy_contracts.sh ${type} && \
    ./3_build_bsn.sh ${type} && \
    ./4_build_config.sh && \
    ./5_build_nodes.sh  ${type}
fi









