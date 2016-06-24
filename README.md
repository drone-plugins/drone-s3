# drone-s3

[![Build Status](http://beta.drone.io/api/badges/drone-plugins/drone-s3/status.svg)](http://beta.drone.io/drone-plugins/drone-s3)
[![Image Size](https://badge.imagelayers.io/plugins/s3:latest.svg)](https://imagelayers.io/?images=plugins/s3:latest 'Get your own badge on imagelayers.io')

Drone plugin to publish files and artifacts to Amazon S3. For the usage information and a listing of the available options please take a look at [the docs](DOCS.md).

## Build

Build the binary with the following commands:

```
export GO15VENDOREXPERIMENT=1
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

go build -a -tags netgo
```

## Docker

Build the docker image with the following commands:

```
docker build --rm=true -t plugins/s3 .
```

Please note incorrectly building the image for the correct x64 linux and with GCO disabled will result in an error when running the Docker image:

```
docker: Error response from daemon: Container command
'/bin/drone-s3' not found or does not exist..
```

## Usage

Build and publish from your current working directory:

```
docker run --rm                     \
  -e PLUGIN_SOURCE=<source>         \
  -e PLUGIN_TARGET=<target>         \
  -e PLUGIN_BUCKET=<bucket>         \
  -e AWS_ACCESS_KEY_ID=<token>      \
  -e AWS_SECRET_ACCESS_KEY=<secret> \
  -v $(pwd):$(pwd)                  \
  -w $(pwd)                         \
  plugins/s3 --dry-run
```
