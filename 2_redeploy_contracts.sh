docker-compose stop && \
docker-compose up -d ganache_net1 && \
docker-compose up -d ganache_net2 && \
make -C external/eth-contracts
make -C external/eth-contracts eth-local-migrate
