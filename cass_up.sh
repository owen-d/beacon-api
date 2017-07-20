#!/bin/bash

set -e

docker run --name cass -d -p 9042:9042 --rm cassandra:3.10

sleep 20

./load_cassandra.sh
