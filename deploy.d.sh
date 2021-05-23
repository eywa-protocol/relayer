#!/usr/bin/env bash

#
# Prerequisites
# - bootstrap exist (earlier has done truffle migration,  make wrappers)
# - binary file was compiled with bootstrap.env

# go build -o bridge  ./cmd/main.go 

export SCALED_NUM=p2p-bridge_node_1 # crutch for ECDSA_KEY_1 / ECDSA_KEY_2
export ECDSA_KEY_1=0x187994e6d93bb29e386fe7ab50232a6a2bea5d6f61046b803b9e9b8306b7d268 # TODO: keystore
export ECDSA_KEY_2=0x3fdb56439eb7c05074586993925c6e06103a5b770b46aa29e399cc693d44ddf7 # TODO: keystore
./bridge -mode init -cnf ./bootstrap.env
./bridge -cnf ./bootstrap.env

