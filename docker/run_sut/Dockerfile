FROM alpine:3.7

RUN apk --update --no-cache add iptables

COPY run.sh  run.sh
RUN chmod +x run.sh

CMD ["./run.sh"]
