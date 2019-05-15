FROM alpine:3.7

RUN apk update              \
 && apk add ca-certificates \
 && apk add iptables


COPY ./cert/server.crt /usr/local/share/ca-certificates/server.crt
RUN update-ca-certificates 2>/dev/null
COPY ./cert /cert/

COPY run.sh  run.sh
RUN chmod +x run.sh

CMD ["./run.sh"]
