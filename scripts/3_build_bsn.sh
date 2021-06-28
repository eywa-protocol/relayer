#!/usr/bin/env bash

touch ../node/bsn.yaml
if [[ "$OSTYPE" == "darwin"* ]];then
  echo "compose bsn for macos docker host"
  # stop and remove bootstrap nodes containers
  docker-compose -f ../docker-compose-macos.yaml rm -s -f bsn1 bsn2 bsn3 && \
  # build image for bootstrap and bridge nodes
  docker-compose -f ../docker-compose-macos.yaml build bsn1 && \
  # init bootstrap nodes
  docker-compose -f ../docker-compose-macos.yaml run --rm  bsn1 ./bsn -mode init
  docker-compose -f ../docker-compose-macos.yaml run --rm  bsn2 ./bsn -mode init
  docker-compose -f ../docker-compose-macos.yaml run --rm  bsn3 ./bsn -mode init
  # up bootstrap nodes
  docker-compose -f ../docker-compose-macos.yaml up -d bsn1 bsn2 bsn3

  sleep 5

  export BSN_URL1=$(docker-compose -f ../docker-compose-macos.yaml exec bsn1 cat keys/bootstrap-peer.env)
  export BSN_URL2=$(docker-compose -f ../docker-compose-macos.yaml exec bsn2 cat keys/bootstrap-peer.env)
  export BSN_URL3=$(docker-compose -f ../docker-compose-macos.yaml exec bsn3 cat keys/bootstrap-peer.env)
else
  echo "compose bsn for linux docker host"
  # stop and remove bootstrap nodes containers
  docker-compose -f ../docker-compose.yaml rm -s -f bsn1 bsn2 bsn3 && \
  # build image for bootstrap and bridge nodes
  docker-compose -f ../docker-compose.yaml build bsn1 && \
  # init bootstrap nodes
  docker-compose -f ../docker-compose.yaml run --rm  bsn1 ./bsn -mode init
  docker-compose -f ../docker-compose.yaml run --rm  bsn2 ./bsn -mode init
  docker-compose -f ../docker-compose.yaml run --rm  bsn3 ./bsn -mode init
  # up bootstrap nodes
  docker-compose -f ../docker-compose.yaml up -d bsn1 bsn2 bsn3

  sleep 5

  export BSN_URL1=$(docker-compose -f ../docker-compose.yaml exec bsn1 cat keys/bootstrap-peer.env)
  export BSN_URL2=$(docker-compose -f ../docker-compose.yaml exec bsn2 cat keys/bootstrap-peer.env)
  export BSN_URL3=$(docker-compose -f ../docker-compose.yaml exec bsn3 cat keys/bootstrap-peer.env)
fi

# build shared bootstrap nodes config for use in bridge nodes
cat > ../.data/bsn.yaml <<EOF
bootstrap-addrs:
  - "${BSN_URL1}"
  - "${BSN_URL2}"
  - "${BSN_URL3}"
EOF