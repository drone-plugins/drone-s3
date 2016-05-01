# Docker image for the Drone build runner
#
#     CGO_ENABLED=0 go build -a -tags netgo
#     docker build --rm=true -t plugins/drone-s3 .

FROM gliderlabs/alpine:3.1
RUN apk add --update \
	python \
	py-pip \
	&& pip install awscli \
	&& aws configure set default.s3.signature_version s3v4
ADD drone-s3 /bin/
ENTRYPOINT ["/bin/drone-s3"]
