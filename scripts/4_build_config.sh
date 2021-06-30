#!/usr/bin/env bash
cat > ../.data/bridge.yaml <<EOF
ticker_interval: 10s
EOF
cat >> ../.data/bridge.yaml <<EOF
chains:
EOF
for f in ../external/eth-contracts/truffle/networks_env/*.env
do

    set -o allexport
      source $f
    set +o allexport
    if [ "$NETWORK_ID" == "1111" ];then
      export ECDSA_KEY=0x60cc6f7a3b09e5080dc86cc0fd80e29545683ad4336012b221998b448d2d57bb
    elif [ "$NETWORK_ID" == "1112" ];then
      export ECDSA_KEY=0x72a21bf9881eff3ab6e4bd245074cdbb6631388e22318bd4f5df2752b69222cc
    elif [ "$NETWORK_ID" == "1113" ];then
      export ECDSA_KEY=0x3fdb56439eb7c05074586993925c6e06103a5b770b46aa29e399cc693d44ddf7
    fi
    cat >> ../.data/bridge.yaml <<EOF
  - id: $NETWORK_ID
    ecdsa_key: $ECDSA_KEY
    bridge_address: $BRIDGE_ADDRESS
    node_list_address: $NODELIST_ADDRESS
    dex_pool_address: $DEXPOOL_ADDRESS
    rpc_urls:
      - $RPC_URL
EOF
done
cat ../.data/bsn.yaml >> ../.data/bridge.yaml






