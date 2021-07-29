#!/usr/bin/env bash
./fix_docker_perms.sh && \
./data.sh && \
./1_clean_nodes.sh && \
./2_redeploy_contracts.sh && \
./fix_docker_perms.sh && \
./3_build_bsn.sh && \
./4_build_config.sh && \
./5_build_nodes.sh




