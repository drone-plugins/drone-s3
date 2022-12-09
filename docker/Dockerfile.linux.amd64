FROM alpine:3.17 as alpine
RUN apk add -U --no-cache ca-certificates

FROM scratch

LABEL maintainer="Drone.IO Community <drone-dev@googlegroups.com>" \
  org.label-schema.name="Drone S3" \
  org.label-schema.vendor="Drone.IO Community" \
  org.label-schema.schema-version="1.0"

ADD release/linux/amd64/drone-s3 /bin/
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/bin/drone-s3"]
