FROM alpine:3.7

RUN apk update              \
 && apk add ca-certificates \
 && apk add iptables


COPY docker_entrypoint.sh  docker_entrypoint.sh
RUN chmod +x docker_entrypoint.sh

CMD ["./docker_entrypoint.sh"]
