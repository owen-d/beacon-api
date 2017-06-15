#!/bin/bash

set -e

KEYSPACE_STR="$(cat <<EOF
CREATE KEYSPACE IF NOT EXISTS bkn
WITH REPLICATION = { 
    'class' : 'SimpleStrategy', 
    'replication_factor' : 2 
};
EOF
)"


cat <(find ./cass/ -type f | xargs cat <(echo $KEYSPACE_STR)) | cqlsh --cqlversion=3.4.4

