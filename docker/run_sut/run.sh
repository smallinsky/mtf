#!/usr/bin/env sh

DOCKER_HOST=$(nslookup host.docker.internal 2> /dev/null | grep Address | cut -d":" -f2 | tr -d " ")

iptables -t nat -A OUTPUT -p tcp --dport 80 -j DNAT --to-destination ${DOCKER_HOST}:8080
iptables -t nat -A OUTPUT -p tcp --dport 443 -j DNAT --to-destination ${DOCKER_HOST}:8443

/component/echo

