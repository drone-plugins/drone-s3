#!/bin/sh

set -e
set -x

GOOS=linux GOARCH=amd64 CGO_ENABLED=0         go build -ldflags "-X main.build=${DRONE_BUILD_NUMBER}" -a -tags netgo -o release/linux/amd64/drone-s3
GOOS=linux GOARCH=arm64 CGO_ENABLED=0         go build -ldflags "-X main.build=${DRONE_BUILD_NUMBER}" -a -tags netgo -o release/linux/arm64/drone-s3
GOOS=linux GOARCH=arm   CGO_ENABLED=0 GOARM=7 go build -ldflags "-X main.build=${DRONE_BUILD_NUMBER}" -a -tags netgo -o release/linux/arm/drone-s3

