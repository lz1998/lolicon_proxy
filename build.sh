#!/bin/bash


GOOS="linux"
GOARCH="amd64"

OUTPUT_FILE="output/lolicon_proxy-${GOOS}-${GOARCH}"
echo $GOOS
echo $GOARCH

GOOS=${GOOS} GOARCH=${GOARCH} go build -v -o ${OUTPUT_FILE}
#GOOS=${GOOS} GOARCH=${GOARCH} CC=x86_64-linux-musl-gcc CGO_ENABLED=1 go build -a -v -o ${OUTPUT_FILE}