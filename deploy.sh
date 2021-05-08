#!/usr/bin/env bash

function parseMyConfig() {
  sed -E "s/$1 *= *([^ ]+).*/\1/;t;d" "$2"
}

function writeMyConfig() {
  if grep -q "$1" "$3"; then
    sed -i "s/$1.*/$1=$2/" "$3"
  else
    echo "$1=$2" >> "$3"
  fi
}

docker-compose stop && docker-compose rm && \
docker-compose up -d ganache_net1 && \
docker-compose up -d ganache_net2 && \
cd ../eth-contracts && \
make clean && \
cd ./truffle && \
npx truffle migrate --reset --network network1 && npx truffle migrate --reset --network network2 && \
cd .. && make wrappers && \
pwd
cd ../p2p-bridge && \


Env1Path="../eth-contracts/truffle/env_connect_to_network1.env"
Env2Path="../eth-contracts/truffle/env_connect_to_network2.env"
BootPath="bootstrap.env"

BRIDGE_ADDRESS_NETWORK1=$(parseMyConfig BRIDGE_ADDRESS_NETWORK1 $Env1Path)
writeMyConfig BRIDGE_ADDRESS_NETWORK1 $BRIDGE_ADDRESS_NETWORK1 $BootPath
BRIDGE_ADDRESS_NETWORK2=$(parseMyConfig BRIDGE_ADDRESS_NETWORK2 $Env2Path)
writeMyConfig BRIDGE_ADDRESS_NETWORK2 $BRIDGE_ADDRESS_NETWORK2 $BootPath

PROXY_NETWORK1=$(parseMyConfig PROXY_NETWORK1 $Env1Path)
writeMyConfig PROXY_NETWORK1 $PROXY_NETWORK1 "$BootPath"
PROXY_NETWORK2=$(parseMyConfig PROXY_NETWORK2 $Env2Path)
writeMyConfig PROXY_NETWORK2 $PROXY_NETWORK2 "$BootPath"

PROXY_ADMIN_NETWORK1=$(parseMyConfig PROXY_ADMIN_NETWORK1 $Env1Path)
writeMyConfig PROXY_ADMIN_NETWORK1 $PROXY_ADMIN_NETWORK1 "$BootPath"
PROXY_ADMIN_NETWORK2=$(parseMyConfig PROXY_ADMIN_NETWORK2 $Env2Path)
writeMyConfig PROXY_ADMIN_NETWORK2 $PROXY_ADMIN_NETWORK2 "$BootPath"

NODELIST_NETWORK1=$(parseMyConfig NODELIST_NETWORK1 $Env1Path)
writeMyConfig NODELIST_NETWORK1 $NODELIST_NETWORK1 "$BootPath"
NODELIST_NETWORK2=$(parseMyConfig NODELIST_NETWORK2 $Env2Path)
writeMyConfig NODELIST_NETWORK2 $NODELIST_NETWORK2 "$BootPath"

MOCKDEX_NETWORK1=$(parseMyConfig MOCKDEX_NETWORK1 $Env1Path)
writeMyConfig MOCKDEX_NETWORK1 $MOCKDEX_NETWORK1 "$BootPath"
MOCKDEX_NETWORK2=$(parseMyConfig MOCKDEX_NETWORK2 $Env2Path)
writeMyConfig MOCKDEX_NETWORK2 $MOCKDEX_NETWORK2 "$BootPath"




account1="0x60cc6f7a3b09e5080dc86cc0fd80e29545683ad4336012b221998b448d2d57bb"
account2="0x187994e6d93bb29e386fe7ab50232a6a2bea5d6f61046b803b9e9b8306b7d268"
account3="0xf5e350eb75d845a8cd538e06331ee6eeb159c54710c6c84c725cd77e2e0dde59"
account4="0x9894defbe159c1abba7db3f88b122bf94c6838a28a98af9466e85c8f573c43bc"
account5="0xafd8124dd3abec91d07dca54878cd296666a96decd122cc7646e9357df21a6de"


account21="0x72a21bf9881eff3ab6e4bd245074cdbb6631388e22318bd4f5df2752b69222cc"
account22="0x3fdb56439eb7c05074586993925c6e06103a5b770b46aa29e399cc693d44ddf7"
account23="0xe84538a8b4059da88d5a76544329093da99603fe05b0f1332f737c74253688c3"
account24="0x469e5c05e289274dd8570c31f2f0f21236f2e071613ac9c565821985e7ae641e"
account25="0x95472b385de2c871fb293f07e76a56e8e93ea4e743fe940afbd44c30730211dc"

make clean
docker-compose up -d --build --scale node=7 --no-recreate
docker-compose ps
docker-compose logs -f -t | grep -v ganache
docker restart $(docker ps -a --format "{{.Names}}" | grep node)




