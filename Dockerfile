# Docker image for the Drone build runner
#
#     CGO_ENABLED=0 go build -a -tags netgo
#     docker build --rm=true -t plugins/s3 .

FROM alpine:3.3

ENV GOPATH /root/go
ENV CGO_ENABLED 0
ENV GO15VENDOREXPERIMENT 1
ENV PKG org/user/drone-s3

ADD vendor $GOPATH/src/$PKG/vendor
ADD *.go $GOPATH/src/$PKG/

RUN apk update && \
    apk add ca-certificates mailcap go && \
    go build -a -tags netgo -o /bin/drone-s3 $PKG && \
    apk del go && \
    rm -rf $GOPATH && \
    rm -rf /var/cache/apk/* && \
    echo "built drone-s3"

ENTRYPOINT ["/bin/drone-s3"]
