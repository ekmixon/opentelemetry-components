#!/bin/bash

# From: https://forums.couchbase.com/t/how-to-create-preconfigured-couchbase-docker-image-with-data/17004

# Enables job control
set -m

# Enables error propagation
set -e

# Run the server and send it to the background
/entrypoint.sh couchbase-server &

# Check if couchbase server is up
check_db() {
  curl --silent http://127.0.0.1:8091/pools > /dev/null
  echo $?
}

# Variable used in echo
i=1
# Echo with
log() {
  echo "[$i] [$(date +"%T")] $@"
  i=`expr $i + 1`
}

# Wait until it's ready
until [[ $(check_db) = 0 ]]; do
  >&2 log "Waiting for Couchbase Server to be available ..."
  sleep 1
done

# Setup index and memory quota
log "$(date +"%T") Init cluster ........."
couchbase-cli cluster-init -c 127.0.0.1 --cluster-username otelu --cluster-password otelpassword \
  --cluster-name otelc --cluster-ramsize 512 --cluster-index-ramsize 256 --services data,index,query,fts \
  --index-storage-setting default

# Create the buckets
log "$(date +"%T") Create buckets ........."
couchbase-cli bucket-create -c 127.0.0.1 --username otelu --password otelpassword --bucket-type couchbase --bucket-ramsize 256 --bucket otelb

# Create the buckets
log "$(date +"%T") Create buckets ........."
couchbase-cli bucket-create -c 127.0.0.1 --username otelu --password otelpassword --bucket-type couchbase --bucket-ramsize 256 --bucket test_bucket

# Create user
log "$(date +"%T") Create users ........."
couchbase-cli user-manage -c 127.0.0.1:8091 -u otelu -p otelpassword --set --rbac-username sysadmin --rbac-password otelpassword \
 --rbac-name "sysadmin" --roles admin --auth-domain local

couchbase-cli user-manage -c 127.0.0.1:8091 -u otelu -p otelpassword --set --rbac-username admin --rbac-password otelpassword \
 --rbac-name "admin" --roles bucket_full_access[*] --auth-domain local

# Need to wait until query service is ready to process N1QL queries
log "$(date +"%T") Waiting ........."
sleep 5

# Create otelb indexes
echo "$(date +"%T") Create otelb indexes ........."
cbq -u otelu -p otelpassword -s "CREATE PRIMARY INDEX idx_primary ON \`otelb\`;"

# Create otelb indexes
echo "$(date +"%T") Create test_bucket indexes ........."
cbq -u otelu -p otelpassword -s "CREATE PRIMARY INDEX idx_primary ON \`test_bucket\`;"

exit 0