FROM alpine:3.4

RUN apk update \
 && apk add vsftpd

COPY fswatch/fswatch /bin
COPY vsftpd.conf /etc/vsftpd/
COPY docker_entrypoint.sh  docker_entrypoint.sh
RUN chmod +x docker_entrypoint.sh

ENTRYPOINT ["./docker_entrypoint.sh"]
