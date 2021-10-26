#!/usr/bin/env bash

ENDPOINT="http://127.0.0.1:15672"

setup_mq() {
    curl -i -u dev:dev \
        -H "content-type:application/json" \
        -X PUT "${ENDPOINT}/api/exchanges/dev/webex" \
        -d'{"type":"direct","auto_delete":false,"durable":true,"internal":false,"arguments":{}}' || return 1


    curl -i -u dev:dev \
        -H "content-type:application/json" \
        -X PUT "${ENDPOINT}/api/queues/dev/webq1" \
        -d'{"auto_delete":false,"durable":true,"arguments":{}}' || return 1


    curl -i -u dev:dev \
        -H "content-type:application/json" \
        -X POST "${ENDPOINT}/api/bindings/dev/e/webex/q/webq1" \
        -d'{"routing_key":"webq1","arguments":{}}' || return 1
}

add_monitor_user() {
    rabbitmq-plugins enable rabbitmq_management
    rabbitmqctl add_user "otelu" "otelp"
    rabbitmqctl set_user_tags "otelu" monitoring
    rabbitmqctl set_permissions -p "dev" "otelu" "" "" ".*"
}

echo "Configuring RabbitMQ. . ."
end=$((SECONDS+60))
while [ $SECONDS -lt $end ]; do
    result="$?"
    if setup_mq; then
        echo "RabbitMQ configured"
        add_monitor_user
        exit 0
    fi
    echo "Trying again in 10 seconds. . ."
    sleep 10
done

echo "Failed to configure permissions"
