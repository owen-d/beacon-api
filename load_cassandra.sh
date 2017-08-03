#!/bin/bash

CQL_VERSION=${CQL_VERSION-3.4.4}
CQL_ARGS="--cqlversion=$CQL_VERSION --connect-timeout=30"
set -e

#Keyspace is inlined as it must be called first
KEYSPACE_STR="$(cat <<EOF
CREATE KEYSPACE IF NOT EXISTS bkn
WITH REPLICATION = { 
    'class' : 'SimpleStrategy', 
    'replication_factor' : 1 
};
EOF
)"

# idempotent structure creation
cat <(find ./cass/structure -type f | xargs cat <(echo $KEYSPACE_STR)) | cqlsh $CQL_ARGS
# unless NO_DATA var is present, inject data as well
if [[ -z $NO_DATA ]];
    then
        cat <(find ./cass/data -type f | xargs cat) | cqlsh $CQL_ARGS
    fi



