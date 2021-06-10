#!/usr/bin/env bash

if [[ "$OSTYPE" == "darwin"* ]];then
  if [[ "$(ifconfig lo0 | grep 172.20)" == "" ]];then
    echo "run sudo ./macos_add_interfaces.sh before deploy"
    exit 1
  fi
  echo "compose for macos docker host"
  docker-compose -f ../docker-compose-macos.yaml stop && \
  docker-compose -f ../docker-compose-macos.yaml up -d --build ganache_net1 && \
  docker-compose -f ../docker-compose-macos.yaml up -d --build ganache_net2 && \
  make -C ../external/eth-contracts && \
  make -C ../external/eth-contracts eth-local-migrate
else
  echo "compose for linux docker host"
  docker-compose stop && \
  docker-compose up -d --build ganache_net1 && \
  docker-compose up -d --build ganache_net2 && \
  make -C ../external/eth-contracts && \
  make -C ../external/eth-contracts eth-local-migrate
fi
