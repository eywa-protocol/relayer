#!/usr/bin/env bash
if [[ "$OSTYPE" == "linux"* ]];then
  sudo chown -R $USER:$USER ../.data/
fi
