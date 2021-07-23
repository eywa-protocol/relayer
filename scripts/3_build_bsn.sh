#!/usr/bin/env bash

touch ../.env.sentry

if [[ "$OSTYPE" == "darwin"* ]];then
  DC="../docker-compose-macos.yaml"
else
  DC="../docker-compose.yaml"
fi

echo "compose bsn for macos docker host"
# stop and remove bootstrap nodes containers
docker-compose -f $DC rm -s -f bsn1 bsn2 bsn3 && \
# build image for bootstrap and bridge nodes
docker-compose -f $DC build bsn1 && \
# init bootstrap nodes
docker-compose -f $DC run --rm  bsn1 ./bsn -init
docker-compose -f $DC run --rm  bsn2 ./bsn -init
docker-compose -f $DC run --rm  bsn3 ./bsn -init
# up bootstrap nodes
docker-compose -f $DC up -d bsn1 bsn2 bsn3

sleep 5

export BSN_URL1=$(docker-compose -f ../docker-compose-macos.yaml exec bsn1 cat keys/bootstrap-peer.env)
export BSN_URL2=$(docker-compose -f ../docker-compose-macos.yaml exec bsn2 cat keys/bootstrap-peer.env)
export BSN_URL3=$(docker-compose -f ../docker-compose-macos.yaml exec bsn3 cat keys/bootstrap-peer.env)

# build shared bootstrap nodes config for use in bridge nodes
cat > ../.data/bsn.yaml <<EOF
bootstrap-addrs:
  - "${BSN_URL1}"
  - "${BSN_URL2}"
  - "${BSN_URL3}"
EOF