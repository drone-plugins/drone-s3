# drone-s3

[![Build Status](http://cloud.drone.io/api/badges/drone-plugins/drone-s3/status.svg)](http://cloud.drone.io/drone-plugins/drone-s3)
[![Gitter chat](https://badges.gitter.im/drone/drone.png)](https://gitter.im/drone/drone)
[![Join the discussion at https://discourse.drone.io](https://img.shields.io/badge/discourse-forum-orange.svg)](https://discourse.drone.io)
[![Drone questions at https://stackoverflow.com](https://img.shields.io/badge/drone-stackoverflow-orange.svg)](https://stackoverflow.com/questions/tagged/drone.io)
[![](https://images.microbadger.com/badges/image/plugins/s3.svg)](https://microbadger.com/images/plugins/s3 "Get your own image badge on microbadger.com")
[![Go Doc](https://godoc.org/github.com/drone-plugins/drone-s3?status.svg)](http://godoc.org/github.com/drone-plugins/drone-s3)
[![Go Report](https://goreportcard.com/badge/github.com/drone-plugins/drone-s3)](https://goreportcard.com/report/github.com/drone-plugins/drone-s3)

Drone plugin to publish files and artifacts to Amazon S3 or Minio. For the
usage information and a listing of the available options please take a look at
[the docs](http://plugins.drone.io/drone-plugins/drone-s3/).

Run the following script to install git-leaks support to this repo. Testing PR Trigger.
```
chmod +x ./git-hooks/install.sh
./git-hooks/install.sh
```

## Build

Build the binary with the following commands:

```
go build
go test
```

## Docker

Build the Docker image with the following commands:

```
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -o release/linux/amd64/drone-s3
docker build --rm=true -t plugins/s3 .
```

Please note incorrectly building the image for the correct x64 linux and with
CGO disabled will result in an error when running the Docker image:

```
docker: Error response from daemon: Container command
'/bin/drone-s3' not found or does not exist..
```

## Usage

Execute from the working directory:

* For Upload
```
docker run --rm \
  -e PLUGIN_SOURCE=<source> \
  -e PLUGIN_TARGET=<target> \
  -e PLUGIN_BUCKET=<bucket> \
  -e AWS_ACCESS_KEY_ID=<token> \
  -e AWS_SECRET_ACCESS_KEY=<secret> \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  plugins/s3 --dry-run
```

* For Download
```
docker run --rm \
  -e PLUGIN_SOURCE=<source directory to be downloaded from bucket> \
  -e PLUGIN_BUCKET=<bucket> \
  -e AWS_ACCESS_KEY_ID=<token> \
  -e AWS_SECRET_ACCESS_KEY=<secret> \
  -e PLUGIN_REGION=<region where the bucket is deployed> \
  -e PLUGIN_DOWNLOAD="true" \
  -v $(pwd):$(pwd) \
  -w $(pwd) \
  plugins/s3 --dry-run
```

## Configuration Variables for Secondary Role Assumption with External ID

The following environment variables enable the plugin to assume a secondary IAM role using IRSA, with an External ID if required by the role’s trust policy.

### Variables

#### `PLUGIN_USER_ROLE_ARN`

- **Type**: String
- **Required**: No
- **Description**: Specifies the secondary IAM role to be assumed by the plugin, allowing it to inherit permissions associated with this role and access specific AWS resources.

#### `PLUGIN_USER_ROLE_EXTERNAL_ID`

- **Type**: String
- **Required**: No
- **Description**: Provide the External ID necessary for the role assumption process if the secondary role’s trust policy mandates it. This is often required for added security, ensuring that only authorized entities assume the role.

### Usage Notes

- If the role secondary role (`PLUGIN_USER_ROLE_ARN`) requires an External ID then pass it through `PLUGIN_USER_ROLE_EXTERNAL_ID`.