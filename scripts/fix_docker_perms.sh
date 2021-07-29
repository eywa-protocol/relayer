#!/usr/bin/env bash
DIRECTORY="../.data"

if [[ -d "$DIRECTORY" ]];then
	rm -rf $DIRECTORY
	mkdir -p $DIRECTORY
fi

if [[ "$OSTYPE" == "linux"* ]];then
  sudo chown -R $USER:$USER ../.data/
fi
