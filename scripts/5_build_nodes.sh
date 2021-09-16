#!/usr/bin/env bash

touch ../.env.sentry
touch ../.env.prom


if [[ "$OSTYPE" == "darwin"* ]];then
  DC="../docker-compose-macos.yaml"
  gsn='gsn1'
elif [ "${1}" == "testnet" ];then
  DC="../docker-compose-testnet.yaml"
  gsn=''
else
  DC="../docker-compose.yaml"
  gsn='gsn1'
fi
  # stop and remove nodes
  docker-compose -f $DC rm -f -s $gsn  node1 node2 node3 node4 node5 node6 node7 &&\
  # init nodes
  docker-compose -f $DC run --rm --no-deps  node1 ./bridge -init -cnf bridge.yaml &&\
  docker-compose -f $DC run --rm --no-deps  node2 ./bridge -init -cnf bridge.yaml &&\
  docker-compose -f $DC run --rm --no-deps  node3 ./bridge -init -cnf bridge.yaml &&\
  docker-compose -f $DC run --rm --no-deps  node4 ./bridge -init -cnf bridge.yaml &&\
  docker-compose -f $DC run --rm --no-deps  node5 ./bridge -init -cnf bridge.yaml &&\
  docker-compose -f $DC run --rm --no-deps  node6 ./bridge -init -cnf bridge.yaml &&\
  docker-compose -f $DC run --rm --no-deps  node7 ./bridge -init -cnf bridge.yaml &&\

  # up prometheus
  set -o allexport
    source ../.env.prom
  set +o allexport
  if [ "$PROM_LISTEN_PORT" != "" ];then
    docker-compose -f $DC up -d --no-deps prometheus
  fi

  # up nodes containers
  docker-compose -f $DC up -d --no-deps $gsn node1 node2 node3 node4 node5 node6 node7 && \
  docker start $(docker ps -f "status=exited" --format "{{.Names}}" | grep node)
  # display logs
  docker-compose -f $DC logs -f -t | grep -v ganache
