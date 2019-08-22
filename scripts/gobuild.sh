#!/bin/bash

export GOOS=$BUILD_OS
export GOARCH=$BUILD_ARCH
export CGO=0

if [ "$BUILD_CGO" != "" ]; then
	export CGO=1
fi

go build -ldflags "-s -X $BUILD_VERSION" -tags "$BUILD_TAGS" -o "$BUILD_OUTDIR/$BUILD_OUTFILE" "$BUILD_APP"
if [ "$?" != "0" ]; then
	echo "Failed!";
	exit 1;
fi

