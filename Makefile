.RECIPEPREFIX = >
.PHONY: _pwd_prompt decrypt_conf encrypt_conf deploy


# CONF_FILE=conf/config.json
CONF_DIR=conf/settings
ENCRYPTED_FILE=conf.aes
HELM_SECRET_NAME=v1api-configs
HELM_NAMESPACE=api

# 'private' task for echoing instructions

_pwd_prompt:
> @echo "Contact ow.diehl@gmail.com for the password."

# decrypt, process tarball & drop in dir
decrypt_conf: _pwd_prompt
> mkdir -p conf/settings 2>/dev/null || :
> openssl aes-256-cbc -d -in ${ENCRYPTED_FILE} | tar -xzv -C ${CONF_DIR}
> chmod -R 700 ${CONF_DIR}

# for updating conf/settings.json
# compress tarball, pass to encrypt
encrypt_conf: _pwd_prompt
> tar -cz -C ${CONF_DIR} . | openssl aes-256-cbc -e -out ${ENCRYPTED_FILE}

# for deploying via helm
deploy: decrypt_conf
> cd conf && ./create_secret.sh ${HELM_NAMESPACE} ${HELM_SECRET_NAME}
> HELM_SECRET_HASH=`find ${CONF_DIR} -type f | xargs cat | shasum -a 256 | awk '{print $$1}'` ; \
> cd k8s ; \
> helm upgrade --install --namespace api --values ./extravals.yaml v1api ./sharecrows-api \
> --set api.configs.secretName=${HELM_SECRET_NAME} --set api.configs.secretHash=$$HELM_SECRET_HASH
