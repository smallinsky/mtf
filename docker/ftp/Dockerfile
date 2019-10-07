FROM golang:1.13.1-alpine3.10 AS builder

RUN apk add git
RUN go get github.com/smallinsky/mtf/cmd/fswatch

FROM alpine:3.10

RUN apk update \
 && apk add vsftpd

COPY --from=builder /go/bin/fswatch /go/bin/
COPY vsftpd.conf /etc/vsftpd/
COPY docker_entrypoint.sh  docker_entrypoint.sh
RUN chmod +x docker_entrypoint.sh

ENTRYPOINT ["./docker_entrypoint.sh"]
