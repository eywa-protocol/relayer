#!/usr/bin/env bash

touch ../node/bsn.yaml
if [[ "$OSTYPE" == "darwin"* ]];then
  echo "compose bsn for macos docker host"

  docker-compose -f ../docker-compose-macos.yaml stop bsn1 bsn2 bsn3 && \
  docker-compose -f ../docker-compose-macos.yaml up -d --build --no-deps bsn1 && \
  docker-compose -f ../docker-compose-macos.yaml up -d bsn2 && \
  docker-compose -f ../docker-compose-macos.yaml up -d bsn3

  sleep 5

  export BSN_URL1=$(docker-compose -f ../docker-compose-macos.yaml exec bsn1 cat keys/bootstrap-peer.env)
  export BSN_URL2=$(docker-compose -f ../docker-compose-macos.yaml exec bsn2 cat keys/bootstrap-peer.env)
  export BSN_URL3=$(docker-compose -f ../docker-compose-macos.yaml exec bsn3 cat keys/bootstrap-peer.env)
else
  echo "compose bsn for linux docker host"

  docker-compose -f ../docker-compose.yaml stop bsn1 bsn2 bsn3 && \
  docker-compose -f ../docker-compose.yaml up -d --build --no-deps bsn1 && \
  docker-compose -f ../docker-compose.yaml up -d bsn2 && \
  docker-compose -f ../docker-compose.yaml up -d bsn3

  sleep 5

  export BSN_URL1=$(docker-compose -f ../docker-compose.yaml exec bsn1 cat keys/bootstrap-peer.env)
  export BSN_URL2=$(docker-compose -f ../docker-compose.yaml exec bsn2 cat keys/bootstrap-peer.env)
  export BSN_URL3=$(docker-compose -f ../docker-compose.yaml exec bsn3 cat keys/bootstrap-peer.env)
fi

cat > ../bsn.yaml <<EOF
bootstrap-addrs:
  - "${BSN_URL1}"
  - "${BSN_URL2}"
  - "${BSN_URL3}"
EOF