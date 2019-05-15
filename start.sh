#!/bin/bash
#need envs APP=proxy/agent, PORT= runging port, APP_PORT=host port for docker host, HOST_IP
#PROXYSERVER, SERVERS=oauth,common-files, REGISTRY
app=$APP
config=/etc/jwaoo/$app.yaml
create_config()
{
	echo "port: $PORT" > $config
	echo "proxy:" >> $config
	echo "  server: $PROXYSERVER" >> $config
	echo "  register: /api/v1/register" >> $config
	echo "  keepalive: /api/v1/keepalive" >> $config
	echo "  services: [$SERVERS]" >> $config
	echo "  beatheart: 2" >> $config
	echo "docker:" >> $config
	echo "  reposity: $REGISTRY" >> $config
}

create_config

exec /usr/bin/$app