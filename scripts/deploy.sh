#!/usr/bin/env bash
type=${1}
./1_clean_nodes.sh && \
./2_redeploy_contracts.sh ${type} && \
./3_build_bsn.sh ${type} && \
./4_build_config.sh && \
./5_build_nodes.sh  ${type}




