#!/usr/bin/env bash

if [[ "$OSTYPE" == "darwin"* ]];then
  export RANDEVOUE=$(LC_CTYPE=C tr -dc 'a-zA-Z0-9' < /dev/urandom | fold -w ${1:-32} | head -n 1)
  # stop and remove nodes
  docker-compose -f ../docker-compose-macos.yaml rm -f -s node1 node2 node3 node4 node5 node6 node7 &&\
  # init nodes
  docker-compose -f ../docker-compose-macos.yaml run --rm --no-deps  node1 ./bridge -init -cnf bridge.yaml &&\
  docker-compose -f ../docker-compose-macos.yaml run --rm --no-deps  node2 ./bridge -init -cnf bridge.yaml &&\
  docker-compose -f ../docker-compose-macos.yaml run --rm --no-deps  node3 ./bridge -init -cnf bridge.yaml &&\
  docker-compose -f ../docker-compose-macos.yaml run --rm --no-deps  node4 ./bridge -init -cnf bridge.yaml &&\
  docker-compose -f ../docker-compose-macos.yaml run --rm --no-deps  node5 ./bridge -init -cnf bridge.yaml &&\
  docker-compose -f ../docker-compose-macos.yaml run --rm --no-deps  node6 ./bridge -init -cnf bridge.yaml &&\
  docker-compose -f ../docker-compose-macos.yaml run --rm --no-deps  node7 ./bridge -init -cnf bridge.yaml &&\
  # up nodes containers
  docker-compose -f ../docker-compose-macos.yaml up -d --no-deps node1 node2 node3 node4 node5 node6 node7 && \
  docker start $(docker ps -f "status=exited" --format "{{.Names}}" | grep node) &&\
  # display logs
  docker-compose -f ../docker-compose-macos.yaml logs -f -t | grep -v ganache
else
  export RANDEVOUE=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w ${1:-32} | head -n 1)
  # stop and remove nodes
  docker-compose -f ../docker-compose.yaml rm -f -s node1 node2 node3 node4 node5 node6 node7
  # init nodes
  docker-compose -f ../docker-compose.yaml run --rm --no-deps  node1 ./bridge -init -cnf bridge.yaml
  docker-compose -f ../docker-compose.yaml run --rm --no-deps  node2 ./bridge -init -cnf bridge.yaml
  docker-compose -f ../docker-compose.yaml run --rm --no-deps  node3 ./bridge -init -cnf bridge.yaml
  docker-compose -f ../docker-compose.yaml run --rm --no-deps  node4 ./bridge -init -cnf bridge.yaml
  docker-compose -f ../docker-compose.yaml run --rm --no-deps  node5 ./bridge -init -cnf bridge.yaml
  docker-compose -f ../docker-compose.yaml run --rm --no-deps  node6 ./bridge -init -cnf bridge.yaml
  docker-compose -f ../docker-compose.yaml run --rm --no-deps  node7 ./bridge -init -cnf bridge.yaml
  # up nodes containers
  docker-compose -f ../docker-compose.yaml up -d --no-deps node1 node2 node3 node4 node5 node6 node7
  docker start $(docker ps -f "status=exited" --format "{{.Names}}" | grep node)
  # display logs
  docker-compose -f ../docker-compose.yaml logs -f -t | grep -v ganache
fi

