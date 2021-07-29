#!/usr/bin/env bash
DIRECTORY="../.data"

if [[ "$OSTYPE" == "linux"* ]] && [[ -d "$DIRECTORY" ]];then
  sudo chown -R $USER:$USER $DIRECTORY
fi
