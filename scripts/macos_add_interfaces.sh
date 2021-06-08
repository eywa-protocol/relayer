#!/usr/bin/env bash

if [[ "$OSTYPE" == "darwin"* ]]; then
  sudo ifconfig lo0 alias 172.20.128.11
  sudo ifconfig lo0 alias 172.20.128.12
fi