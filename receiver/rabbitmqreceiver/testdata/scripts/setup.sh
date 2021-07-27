#!/usr/bin/env bash

set -e

ENDPOINT="http://127.0.0.1:15672"

setup_mq() {
    curl -i -u dev:dev \
        -H "content-type:application/json" \
        -X PUT "${ENDPOINT}/api/exchanges/dev/webex" \
        -d'{"type":"direct","auto_delete":false,"durable":true,"internal":false,"arguments":{}}'


    curl -i -u dev:dev \
        -H "content-type:application/json" \
        -X PUT "${ENDPOINT}/api/queues/dev/webq1" \
        -d'{"auto_delete":false,"durable":true,"arguments":{}}'


    curl -i -u dev:dev \
        -H "content-type:application/json" \
        -X POST "${ENDPOINT}/api/bindings/dev/e/webex/q/webq1" \
        -d'{"routing_key":"webq1","arguments":{}}'
}

echo "Configuring RabbitMQ. . ."
end=$((SECONDS+60))
while [ $SECONDS -lt $end ]; do
    result="$?"
    if setup_mq; then
        echo "RabbitMQ configured"
        exit 0
    fi
    echo "Trying again in 5 seconds. . ."
    sleep 5
done

echo "Failed to configure permissions"

