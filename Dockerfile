FROM golang:1.9-alpine
WORKDIR /go/src/github.com/quintoandar/drone-s3
ADD . .
RUN GOOS=linux CGO_ENABLED=0 go build -o /bin/drone-s3 \
    github.com/quintoandar/drone-s3

FROM scratch
COPY --from=0 /bin/drone-s3 /bin/drone-s3
ENTRYPOINT ["/bin/drone-s3"]
