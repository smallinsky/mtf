FROM alpine:3.7

RUN apk --update --no-cache add iptables

COPY service/echo/echo sut
COPY forward.sh  forward.sh
RUN chmod +x forward.sh

CMD ["./forward.sh"]
