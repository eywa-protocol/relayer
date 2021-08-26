#!/usr/bin/env bash

HARDHAT="../external/eth-contracts/hardhat"
EXAMPLE="/.env.example"
CONFIG="/.env"

function copyConfigs() {
  echo $HARDHAT$EXAMPLE $HARDHAT$CONFIG
  cp $HARDHAT$EXAMPLE $HARDHAT$CONFIG
}

copyConfigs

if [[ "$OSTYPE" == "darwin"* ]];then
  export RANDEVOUE=$(LC_CTYPE=C tr -dc 'a-zA-Z0-9' < /dev/urandom | fold -w ${1:-32} | head -n 1)
  export SED=gsed
else
  export RANDEVOUE=$(head -80 /dev/urandom | LC_ALL=c tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)
  export SED=sed
fi

cat > ../.data/bridge.yaml <<EOF
ticker_interval: 10s
rendezvous: $RANDEVOUE
EOF
cat > ../.data/gsn.yaml <<EOF
ticker_interval: 10s
EOF

cat >> ../.data/bridge.yaml <<EOF
chains:
EOF
cat >> ../.data/gsn.yaml <<EOF
chains:
EOF

for f in ../external/eth-contracts/hardhat/networks_env/env_connect_to_*.env
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
    elif [ "$NETWORK_ID" == "4" ];then
      export ECDSA_KEY=0x3fdb56439eb7c05074586993925c6e06103a5b770b46aa29e399cc693d44ddf7  # 0x2b3cc5fcAC62299520FA96D75f125c33B48E70d7
    elif [ "$NETWORK_ID" == "97" ];then
      export ECDSA_KEY=0x3fdb56439eb7c05074586993925c6e06103a5b770b46aa29e399cc693d44ddf7  # 0x2b3cc5fcAC62299520FA96D75f125c33B48E70d7
    elif [ "$NETWORK_ID" == "80001" ];then
      export ECDSA_KEY=0x3fdb56439eb7c05074586993925c6e06103a5b770b46aa29e399cc693d44ddf7  # 0x2b3cc5fcAC62299520FA96D75f125c33B48E70d7
    fi

    cat >> ../.data/bridge.yaml <<EOF
  - id: $NETWORK_ID
    ecdsa_key: $ECDSA_KEY
    bridge_address: $BRIDGE_ADDRESS
    node_registry_address: $NODEREGISTRY_ADDRESS
    dex_pool_address: $DEXPOOL_ADDRESS
    reward_address: $REWARDS_ADDRESS
    forwarder_address: $FORWARDER_ADDRESS
    rpc_urls:
      - $RPC_URL
EOF
    cat >> ../.data/gsn.yaml <<EOF
  - id: $NETWORK_ID
    ecdsa_key: $ECDSA_KEY
    forwarder_address: $FORWARDER_ADDRESS
    rpc_urls:
      - $RPC_URL
EOF

CONFIG_KEY=PRIVATE_KEY_NETWORK${NETWORK_ID: -1}
$SED -i "s/$CONFIG_KEY.*/$CONFIG_KEY=$ECDSA_KEY/" "$HARDHAT$CONFIG"
done

cat ../.data/bsn.yaml >> ../.data/bridge.yaml
cat ../.data/bsn.yaml >> ../.data/gsn.yaml






