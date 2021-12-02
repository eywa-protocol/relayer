#!/bin/bash

while ! [ -d .git ]; do cd ..; done
cd external/eth-contracts/hardhat/

for net in network3 network2 network1
do
    npx hardhat run ./scripts/bridge/updateEpoch.js --network ${net}
done
