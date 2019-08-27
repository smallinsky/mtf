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

function forward_http() {
	local DOCKER_HOST=$(nslookup host.docker.internal 2> /dev/null | grep Address | cut -d":" -f2 | tr -d " ")
	iptables -t nat -A OUTPUT -p tcp --dport 80 -j DNAT --to-destination ${DOCKER_HOST}:8080
	iptables -t nat -A OUTPUT -p tcp --dport 443 -j DNAT --to-destination ${DOCKER_HOST}:8443
}


update_ca
forward_http

mkdir -p ~/.config/gcloud
cat > ~/.config/gcloud/application_default_credentials.json << 'EOF'
{
  "client_id": "foo_client_id",
  "client_secret": "foo_client_secret",
  "refresh_token": "foo_refresh_token",
  "type": "authorized_user",
  "auth_uri": "myauth.google.com",
  "token_uri": "mytoken.google.com"
}
EOF

echo "Run SUT"
/component/$SUT_BINARY_NAME
