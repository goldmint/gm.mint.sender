#!/bin/bash

case "$(uname -s)" in
    Linux*)     machine=linux;;
    Darwin*)    machine=mac;;
    CYGWIN*)    machine=cygwin;;
    MINGW*)     machine=mingw;;
    *)          machine="unknown"
esac

if [ "$machine" == "cygwin" ] || [ "$machine" == "mingw" ]; then
	MOUNT_GOPATH="/${GOPATH//\\/$'/'}"
	MOUNT_PWD="/${PWD//\\/$'/'}"
else
	MOUNT_GOPATH=$GOPATH
	MOUNT_PWD=$(PWD)
fi

if [ "$BUILD_OS" == "windows" ]; then
	if [ "$BUILD_ARCH" == "386" ]; then
		CXX='i686-w64-mingw32-g++'
		CC='i686-w64-mingw32-gcc'
	else
		CXX='x86_64-w64-mingw32-g++'
		CC='x86_64-w64-mingw32-gcc'
	fi
else
	CXX='g++'
	CC='gcc'
fi

docker run \
	--rm \
	-i --name gobuild_with_docker \
	-e "GOOS=$BUILD_OS" \
	-e "GOARCH=$BUILD_ARCH" \
	-e "CGO_ENABLED=1" \
	-e "CXX=$CXX" \
	-e "CC=$CC" \
	-v "$MOUNT_GOPATH":/go \
	-v "$MOUNT_PWD":/app \
	gobuild_with_docker \
	go build -ldflags "-s -X $BUILD_VERSION" -tags "$BUILD_TAGS" -o "$BUILD_OUTDIR/$BUILD_OUTFILE" "$BUILD_APP"