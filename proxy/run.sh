#!/bin/bash

# also allow port injection via docker CMD syntax
LISTEN_PORT=$1
PROXY_PORT=$2

TEMPLATE_FILE=/etc/nginx/proxy.template.conf

# inject ports
echo using LISTEN_PORT ${LISTEN_PORT:=80}
echo using PROXY_PORT ${PROXY_PORT:=8080}

sed "s/{{LISTEN_PORT}}/$LISTEN_PORT/g" $TEMPLATE_FILE \
    | sed "s/{{PROXY_PORT}}/$PROXY_PORT/g" \
    > /etc/nginx/conf.d/proxy.conf

echo -e "using config\n\n"
cat /etc/nginx/conf.d/proxy.conf
echo -e "\n\n"

# run nginx
nginx -g "daemon off;"
