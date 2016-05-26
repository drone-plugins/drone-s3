# Docker image for the Drone build runner
#
#     CGO_ENABLED=0 go build -a -tags netgo
#     docker build --rm=true -t plugins/s3 .

FROM alpine:3.3
RUN apk update && apk add ca-certificates mailcap && rm -rf /var/cache/apk/*
ADD drone-s3 /bin/
ENTRYPOINT ["/bin/drone-s3"]
