#!/usr/bin/env bash

docker-compose -f ../docker-compose.yaml build bsn1
docker-compose -f ../docker-compose.yaml up -d --no-deps node1 node2 node3 node4 node5 node6 node7
