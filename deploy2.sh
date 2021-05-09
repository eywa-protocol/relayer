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




Env1Path="../eth-contracts/truffle/env_connect_to_network1.env"
Env2Path="../eth-contracts/truffle/env_connect_to_network2.env"
BootPath="bootstrap.env"

./gana.sh


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



./compose.sh




