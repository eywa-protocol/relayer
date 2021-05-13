#!/usr/bin/env bash

#
# FOR SINGLE MODE
#

docker-compose stop node
docker-compose rm   node
cd ../eth-contracts
make clean
cd ./truffle
npx truffle migrate -f 2 --to 2 --reset --network rinkeby 
npx truffle migrate -f 3 --to 3 --reset --network bsctestnet
cd .. 
make wrappers



cd ../p2p-bridge
git checkout -- ./*.env
make clean

./3_edit_configs.sh

export MODE=singlenode
export ECDSA_KEY_1=0x187994e6d93bb29e386fe7ab50232a6a2bea5d6f61046b803b9e9b8306b7d268
export ECDSA_KEY_2=0x3fdb56439eb7c05074586993925c6e06103a5b770b46aa29e399cc693d44ddf7
export NETWORK_RPC_1=wss://rinkeby.infura.io/ws/v3/ab95bf9f6dd743e6a8526579b76fe358
export NETWORK_RPC_2=ws://95.217.104.54:8576
docker-compose up -d --build  node

