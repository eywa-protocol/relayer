#!/bin/bash

cd external/eth-contracts/hardhat/
for net in network1 network2 network3
do
    npx hardhat run ./scripts/bridge/updateEpoch.js --network ${net}
done
