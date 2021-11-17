#!/usr/bin/env bash

set -e

export COUCHDB_LISTEN_ADDR="localhost:5984"
export COUCHDB_USER=otelu
export COUCHDB_PASSWORD=otelp
export COUCHDB_DATABASE=oteld

curl -X PUT -u "${COUCHDB_USER}:${COUCHDB_PASSWORD}" \
    "http://${COUCHDB_LISTEN_ADDR}/_users"

curl -X PUT -u "${COUCHDB_USER}:${COUCHDB_PASSWORD}" \
    "http://${COUCHDB_LISTEN_ADDR}/_replicator"

curl -X PUT -u "${COUCHDB_USER}:${COUCHDB_PASSWORD}" \
    "http://${COUCHDB_LISTEN_ADDR}/_global_changes"

curl -X PUT -u "${COUCHDB_USER}:${COUCHDB_PASSWORD}" \
    "http://${COUCHDB_LISTEN_ADDR}/${COUCHDB_DATABASE}"
