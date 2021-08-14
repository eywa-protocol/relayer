#!/usr/bin/env bash

TPATH="./test/networks"

if [[ "${1}" == "local" ]];then
    COMMAND="Test_Local_SendRequestV2"
  else
    COMMAND="Test_SendRequestV2"
fi

go test -v  ${TPATH} -run ${COMMAND}_FromRinkebyToBsc $(date +%s)
go test -v  ${TPATH} -run ${COMMAND}_FromBscToRinkeby $(date +%s)
go test -v  ${TPATH} -run ${COMMAND}_FromRinkebyToBsc $(date +%s)
go test -v  ${TPATH} -run ${COMMAND}_FromBscToRinkeby $(date +%s)
go test -v  ${TPATH} -run ${COMMAND}_FromRinkebyToBsc $(date +%s)
go test -v  ${TPATH} -run ${COMMAND}_FromRinkebyToMumbai $(date +%s)
go test -v  ${TPATH} -run ${COMMAND}_FromRinkebyToBsc $(date +%s)
go test -v  ${TPATH} -run ${COMMAND}_FromBscToRinkeby $(date +%s)
go test -v  ${TPATH} -run ${COMMAND}_FromBscToMumbai $(date +%s)
go test -v  ${TPATH} -run ${COMMAND}_FromRinkebyToBsc $(date +%s)
go test -v  ${TPATH} -run ${COMMAND}_FromRinkebyToBsc $(date +%s)

