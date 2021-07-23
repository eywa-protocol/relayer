#!/usr/bin/env bash

if [[ $1 = "" ]];then
  echo "sentry turned OFF"
    echo "" > ../.env.sentry
else
  echo "sentry turned ON for developer $1"
  sed "s/developer_name/$1/" ../.env.sentry_example >../.env.sentry
fi