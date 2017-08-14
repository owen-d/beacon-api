.RECIPEPREFIX = >
.PHONY: _pwd_prompt decrypt_conf encrypt_conf


# CONF_FILE=conf/config.json
CONF_DIR=k8s/conf/settings
ENCRYPTED_FILE=conf.aes

# 'private' task for echoing instructions

_pwd_prompt:
> @echo "Contact ow.diehl@gmail.com for the password."

# decrypt, process tarball & drop in dir
decrypt_conf: _pwd_prompt
> mkdir -p ${CONF_DIR} 2>/dev/null || :
> openssl aes-256-cbc -d -in ${ENCRYPTED_FILE} | tar -xzv -C ${CONF_DIR}
> chmod -R 700 ${CONF_DIR}

# for updating conf/settings.json
# compress tarball, pass to encrypt
encrypt_conf: _pwd_prompt
> tar -cz -C ${CONF_DIR} . | openssl aes-256-cbc -e -out ${ENCRYPTED_FILE}
