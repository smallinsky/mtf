#!/usr/bin/env sh

function update_ca() {
  while :;
  do
    if [ -f /usr/local/share/ca-certificates/server.crt ]; then
      update-ca-certificates
      echo "Certs updated"
      break
    fi
    sleep 0.1
  done
}

function cp_cert() {
  mkdir -p /tmp/mtf/cert
  cp /usr/local/share/ca-certificates/server.key /tmp/mtf/cert
  cp /usr/local/share/ca-certificates/server.crt /tmp/mtf/cert
}

function forward_http() {
  local DOCKER_HOST=$(nslookup $DOCKER_HOST_ADDR 2> /dev/null | grep Address | cut -d":" -f2 | tr -d " ")
  iptables -t nat -A OUTPUT -p tcp --dport 80 -j DNAT --to-destination ${DOCKER_HOST}:8080
  iptables -t nat -A OUTPUT -p tcp --dport 443 -j DNAT --to-destination ${DOCKER_HOST}:8443
}


update_ca
cp_cert
forward_http

mkdir -p ~/.config/gcloud
cat > ~/.config/gcloud/application_default_credentials.json << 'EOF'
{
  "client_id": "test_client_id",
  "client_secret": "test_client_secret",
  "refresh_token": "test_refresh_token",
  "type": "authorized_user",
  "auth_uri": "myauth.google.com",
  "token_uri": "mytoken.google.com"
}
EOF

echo "Run SUT"
/component/$SUT_BINARY_NAME
