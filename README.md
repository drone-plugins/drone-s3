# drone-s3

[![Build Status](http://beta.drone.io/api/badges/drone-plugins/drone-s3/status.svg)](http://beta.drone.io/drone-plugins/drone-s3)
[![](https://badge.imagelayers.io/plugins/drone-s3:latest.svg)](https://imagelayers.io/?images=plugins/drone-s3:latest 'Get your own badge on imagelayers.io')

Drone plugin to upload files and artifacts to S3

## Usage

```sh
./drone-s3 <<EOF
{
    "repo": {
        "clone_url": "git://github.com/drone/drone",
        "full_name": "drone/drone"
    },
    "build": {
        "event": "push",
        "branch": "master",
        "commit": "436b7a6e2abaddfd35740527353e78a227ddcb2c",
        "ref": "refs/heads/master"
    },
    "workspace": {
        "root": "/drone/src",
        "path": "/drone/src/github.com/drone/drone"
    },
    "vargs": {
    }
}
EOF
```

## Docker

Build the Docker container using `make`:

```sh
make deps build
docker build --rm=true -t plugins/drone-s3 .
```

### Example

```sh
docker run -i plugins/drone-s3 <<EOF
{
    "repo": {
        "clone_url": "git://github.com/drone/drone",
        "full_name": "drone/drone"
    },
    "build": {
        "event": "push",
        "branch": "master",
        "commit": "436b7a6e2abaddfd35740527353e78a227ddcb2c",
        "ref": "refs/heads/master"
    },
    "workspace": {
        "root": "/drone/src",
        "path": "/drone/src/github.com/drone/drone"
    },
    "vargs": {
    }
}
EOF
```
