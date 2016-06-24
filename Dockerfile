# Docker image for the Drone build runner
#
#     CGO_ENABLED=0 go build -a -tags netgo
#     docker build --rm=true -t plugins/s3 .

FROM alpine:3.3

ENV GOPATH /root/go
ENV CGO_ENABLED 0
ENV PKG org/user/drone-s3

RUN apk update && \
    apk add ca-certificates mailcap go && \
    mkdir -p $GOPATH/src/$PKG && \
    mv * $GOPATH/src/$PKG/ && \
    go build -a -tags netgo -o /bin/drone-s3 $PKG && \
    apk del go && \
    rm -rf /var/cache/apk/* && \
    echo "built drone-s3"

ENTRYPOINT ["/bin/drone-s3"]
