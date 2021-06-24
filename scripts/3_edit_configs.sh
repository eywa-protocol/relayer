#!/usr/bin/env bash

function parseMyConfig() {
  if [[ "$OSTYPE" == "darwin"* ]]; then
    gsed -E "s/$1 *= *([^ ]+).*/\1/;t;d" "$2"
  else
    sed -E "s/$1 *= *([^ ]+).*/\1/;t;d" "$2"
  fi
}

function writeMyConfig() {
  if grep -q "$1" "$3"; then
    if [[ "$OSTYPE" == "darwin"* ]]; then
      gsed -i "s/$1.*/$1=$2/" "$3"
    else
      sed -i "s/$1.*/$1=$2/" "$3"
    fi
  else
    echo "$1=$2" >> "$3"
  fi
}



Env1Path="../external/eth-contracts/truffle/env_connect_to_network1.env"
Env2Path="../external/eth-contracts/truffle/env_connect_to_network2.env"
Env3Path="../external/eth-contracts/truffle/env_connect_to_network3.env"

BootPath="../bootstrap.env"



BRIDGE_NETWORK1=$(parseMyConfig BRIDGE_NETWORK1    $Env1Path)
writeMyConfig BRIDGE_NETWORK1    $BRIDGE_NETWORK1    $BootPath
BRIDGE_NETWORK2=$(parseMyConfig BRIDGE_NETWORK2 $Env2Path)
writeMyConfig BRIDGE_NETWORK2 $BRIDGE_NETWORK2 $BootPath
BRIDGE_NETWORK3=$(parseMyConfig BRIDGE_NETWORK3 $Env3Path)
writeMyConfig BRIDGE_NETWORK3 $BRIDGE_NETWORK3 $BootPath


NODELIST_NETWORK1=$(parseMyConfig NODELIST_NETWORK1 $Env1Path)
writeMyConfig NODELIST_NETWORK1 $NODELIST_NETWORK1 "$BootPath"
NODELIST_NETWORK2=$(parseMyConfig NODELIST_NETWORK2 $Env2Path)
writeMyConfig NODELIST_NETWORK2 $NODELIST_NETWORK2 "$BootPath"
NODELIST_NETWORK3=$(parseMyConfig NODELIST_NETWORK3 $Env3Path)
writeMyConfig NODELIST_NETWORK3 $NODELIST_NETWORK3 "$BootPath"









