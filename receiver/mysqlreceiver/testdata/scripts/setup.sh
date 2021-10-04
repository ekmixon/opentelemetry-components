#!/usr/bin/env bash

set -e

USER="otel"
CODE=1


setup_permissions() {
    mysql -u root -e "CREATE USER ${USER}" > /dev/null
    mysql -u root -e "GRANT PROCESS ON *.* TO ${USER}" > /dev/null
    mysql -u root -e "FLUSH PRIVILEGES" > /dev/null
}

echo "Configuring ${USER} permissions. . ."
end=$((SECONDS+60))
while [ $SECONDS -lt $end ]; do
    result="$?"
    if setup_permissions; then
        echo "Permissions configured!"
        exit 0
    fi
    echo "Trying again in 5 seconds. . ."
    sleep 5
done

echo "Failed to configure permissions"
