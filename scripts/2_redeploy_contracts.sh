#!/usr/bin/env bash

touch ../.env.sentry
touch ../.env.prom

if [[ "${1}" == "local" ]];then
  if [[ "$OSTYPE" == "darwin"* ]];then
    if [[ "$(ifconfig lo0 | grep 172.20)" == "" ]];then
      echo "run sudo ./macos_add_interfaces.sh before deploy"
      exit 1
    fi
    echo "compose for macos docker host"
    docker-compose -f ../docker-compose-macos.yaml rm -f -s -v ganache_net1 ganache_net2 ganache_net3 && \
    docker-compose -f ../docker-compose-macos.yaml up -d --no-deps --build --force-recreate ganache_net1 && \
    docker-compose -f ../docker-compose-macos.yaml up -d --no-deps --build --force-recreate ganache_net2 && \
    docker-compose -f ../docker-compose-macos.yaml up -d --no-deps --build --force-recreate ganache_net3 && \
    make -C ../external/eth-contracts eth-local-migrate
  else
    echo "compose for linux docker host"
    docker-compose -f ../docker-compose.yaml stop ganache_net1 ganache_net2 ganache_net3 && \
    docker-compose -f ../docker-compose.yaml up -d --no-deps --build --force-recreate ganache_net1 && \
    docker-compose -f ../docker-compose.yaml up -d --no-deps --build --force-recreate ganache_net2 && \
    docker-compose -f ../docker-compose.yaml up -d --no-deps --build --force-recreate ganache_net3 && \
    make -C ../external/eth-contracts eth-local-migrate
  fi
fi
