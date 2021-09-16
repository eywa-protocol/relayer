#!/usr/bin/env bash

touch ../.env.sentry
touch ../.env.prom


if [[ "$OSTYPE" == "darwin"* ]];then
  DC="../docker-compose-macos.yaml"
else
  DC="../docker-compose.yaml"
fi

# start all containers
docker-compose -f $DC start && \
# display logs
docker-compose -f $DC logs -f -t | grep -v ganache