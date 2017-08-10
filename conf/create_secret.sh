kubectl -n api create secret generic v1api-configs \
        --from-file=./settings/config.json \
        --from-file=./settings/gcp-credentials.json
