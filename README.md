# drone-s3

[![Build Status](http://beta.drone.io/api/badges/drone-plugins/drone-s3/status.svg)](http://beta.drone.io/drone-plugins/drone-s3)
[![Join the chat at https://gitter.im/drone/drone](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/drone/drone)
[![Go Doc](https://godoc.org/github.com/drone-plugins/drone-s3?status.svg)](http://godoc.org/github.com/drone-plugins/drone-s3)
[![Go Report](https://goreportcard.com/badge/github.com/drone-plugins/drone-s3)](https://goreportcard.com/report/github.com/drone-plugins/drone-s3)
[![](https://images.microbadger.com/badges/image/plugins/s3.svg)](https://microbadger.com/images/plugins/s3 "Get your own image badge on microbadger.com")

Drone plugin to publish files and artifacts to Amazon S3 or Minio. For the usage information and a listing of the available options please take a look at [the docs](http://plugins.drone.io/drone-plugins/drone-s3/).

## Build

Build the binary with the following commands:

```
go build
```

## Docker

Build the Docker image with the following commands:

```
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -tags netgo -o release/linux/amd64/drone-s3
docker build --rm -t plugins/s3 .
```

## Usage

Execute from the working directory:

```
docker run --rm \
  -e PLUGIN_DRY_RUN=true \
  -e PLUGIN_SOURCE=<source> \
  -e PLUGIN_TARGET=<target> \
  -e PLUGIN_BUCKET=<bucket> \
  -e AWS_ACCESS_KEY_ID=<token> \
  -e AWS_SECRET_ACCESS_KEY=<secret> \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  plugins/s3
```
