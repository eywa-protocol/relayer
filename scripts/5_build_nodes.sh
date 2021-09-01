#!/usr/bin/env bash

touch ../.env.sentry

if [[ "$OSTYPE" == "darwin"* ]];then
    DC="../docker-compose-macos.yaml"
    NETS="network1 network2 network3"
elif [ "${1}" == "testnet" ];then
    DC="../docker-compose-testnet.yaml"
    NETS="rinkeby bsctestnet mumbai"
else
    DC="../docker-compose.yaml"
    NETS="network1 network2 network3"
fi
# stop and remove nodes
docker-compose -f $DC rm -f -s gsn1 node1 node2 node3 node4 node5 node6 node7
# init nodes
docker-compose -f $DC  up -d --no-deps gsn1

for f in ../external/eth-contracts/hardhat/networks_env/env_connect_to_*.env
do
    set -o allexport
      source $f
    set +o allexport
    if [ "$NETWORK_ID" == "1111" ];then
      NET="network1"
        cd ../external/eth-contracts/hardhat/; ADDR=$FORWARDER_ADDRESS npx hardhat run ./scripts/bridge/mint.js --network ${NET}; cd -
    elif [ "$NETWORK_ID" == "1112" ];then
      NET="network2"
        cd ../external/eth-contracts/hardhat/; ADDR=$FORWARDER_ADDRESS npx hardhat run ./scripts/bridge/mint.js --network ${NET}; cd -
    elif [ "$NETWORK_ID" == "1113" ];then
      NET="network3"
        cd ../external/eth-contracts/hardhat/; ADDR=$FORWARDER_ADDRESS npx hardhat run ./scripts/bridge/mint.js --network ${NET}; cd -
    fi
done

for i in 1 2 3 4 5 6 7; do
    ADDR=$(docker-compose -f $DC run --rm --no-deps node$i ./bridge -init -cnf bridge.yaml -verbosity 0)
    for net in ${NETS}; do
        cd ../external/eth-contracts/hardhat/; ADDR=$ADDR npx hardhat run ./scripts/bridge/mint.js --network ${net}; cd -
    done
    docker-compose -f $DC run --rm --no-deps node$i ./bridge -register -cnf bridge.yaml
done
# up nodes containers
docker-compose -f $DC up -d --no-deps gsn1 node1 node2 node3 node4 node5 node6 node7 && \
    docker start $(docker ps -f "status=exited" --format "{{.Names}}" | grep node)
# display logs
docker-compose -f $DC logs -f -t | grep -v ganache
