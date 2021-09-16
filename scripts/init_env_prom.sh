#!/usr/bin/env bash

if [[ $1 = "" ]];then
  echo "prometheus exporter turned OFF"
    echo "" > ../.env.prom
else
  echo "prometheus exporter turned ON port: $1"
  echo "PROM_LISTEN_PORT=$1" > ../.env.prom
fi