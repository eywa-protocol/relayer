#!/usr/bin/env bash

touch ../.env.sentry
touch ../.env.prom

if [[ "$OSTYPE" == "darwin"* ]];then
  DC="../docker-compose-macos.yaml"
else
  DC="../docker-compose.yaml"
fi

# stop all containers
docker-compose -f $DC stop