docker-compose stop && \
docker-compose up -d ganache_net1 && \
docker-compose up -d ganache_net2 && \
cd ../eth-contracts && \
make && \
cd ./truffle && \
npx truffle migrate --reset --network network1 && npx truffle migrate --reset --network network2
