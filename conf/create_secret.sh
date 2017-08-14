#!/bin/bash
set -e

if [[ -n $1 ]]
then
    SECRET_NAMESPACE=$1
fi

if [[ -n $2 ]]
then
    SECRET_NAME=$2
fi



kubectl create secret generic ${SECRET_NAME:=v1api-configs} \
        --from-file=./settings/config.json \
        --from-file=./settings/gcp-credentials.json \
        --dry-run \
        --save-config \
        -o yaml \
        | kubectl apply -n ${SECRET_NAMESPACE:=api} -f -
