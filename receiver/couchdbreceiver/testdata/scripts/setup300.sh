#!/usr/bin/env bash

set -e

export COUCHDB_LISTEN_ADDR="localhost:5984"
export COUCHDB_USER=otel
export COUCHDB_PASSWORD=otel
export COUCHDB_DATABASE=otel

curl -X PUT -u "${COUCHDB_USER}:${COUCHDB_PASSWORD}" \
    "http://${COUCHDB_LISTEN_ADDR}/${COUCHDB_DATABASE}"