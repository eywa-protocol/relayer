#!/usr/bin/env bash

if [[ "$OSTYPE" == "darwin"* ]]; then
  ifconfig lo0 alias 172.20.128.11
  ifconfig lo0 alias 172.20.128.12
  ifconfig lo0 alias 172.20.128.13
fi