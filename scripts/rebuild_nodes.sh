#!/usr/bin/env bash

touch ../.env.sentry

if [[ "$OSTYPE" == "darwin"* ]];then
  DC="../docker-compose-macos.yaml"
else
  DC="../docker-compose.yaml"
fi

# rebuild image
docker-compose -f $DC build bsn1 && \
# up nodes containers
docker-compose -f $DC up -d --no-deps node1 node2 node3 node4 node5 node6 node7 && \
# display logs
docker-compose -f $DC logs -f -t | grep -v ganache
