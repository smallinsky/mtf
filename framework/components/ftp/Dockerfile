FROM golang:latest AS builder
ADD cmd /go/fswatch
WORKDIR /go/fswatch
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go buld -o fswatch .


FROM alpine:3.4

RUN apk update \
 && apk add vsftpd

COPY --from=builder /go/fswatch/fswatch /bin
COPY vsftpd.conf /etc/vsftpd/
COPY docker_entrypoint.sh  docker_entrypoint.sh
RUN chmod +x docker_entrypoint.sh

ENTRYPOINT ["./docker_entrypoint.sh"]
