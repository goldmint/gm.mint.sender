#!/bin/bash

DATE=`date +%Y%m%d`
if [ "$?" != "0" ]; then
	exit 1
fi

BRANCH=`git rev-parse --abbrev-ref HEAD`
if [ "$?" != "0" ]; then
	exit 1
fi

HASH=`git rev-parse --short HEAD`
if [ "$?" != "0" ]; then
	exit 1
fi

echo $DATE-$HASH-$BRANCH
