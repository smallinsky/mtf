#!/usr/bin/env sh
set -eo

DOCKER_HOST=$(nslookup host.docker.internal | grep Address | cut -d":" -f2 | tr -d " ")
echo "Forwarding http and https traffic to ${DOCKER_HOST}"

iptables -t nat -A OUTPUT -p tcp --dport 80 -j DNAT --to-destination ${DOCKER_HOST}:8080
iptables -t nat -A OUTPUT -p tcp --dport 443 -j DNAT --to-destination ${DOCKER_HOST}:8443

echo "Running SUT"
./sut

