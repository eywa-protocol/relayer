#!/usr/bin/env bash
DIRECTORY="../.data"

if [[ -d "$DIRECTORY" ]];then
	rm -rf $DIRECTORY
fi

mkdir -p $DIRECTORY
