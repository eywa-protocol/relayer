#!/usr/bin/env bash

if [[ "$OSTYPE" == "darwin"* ]]; then
  ifconfig lo0 alias 172.20.128.11
  ifconfig lo0 alias 172.20.128.12
  ifconfig lo0 alias 172.20.128.13
  ifconfig lo0 alias 172.20.64.11
  ifconfig lo0 alias 172.20.64.12
  ifconfig lo0 alias 172.20.64.13
  ifconfig lo0 alias 172.20.30.11
fi