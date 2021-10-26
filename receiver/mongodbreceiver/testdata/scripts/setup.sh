#!/bin/bash
set -e

CODE=1

USER="otelu"
MONGO_INITDB_ROOT_USERNAME="otel"
MONGO_INITDB_ROOT_PASSWORD="otel"

setup_permissions() {
  mongo -u $MONGO_INITDB_ROOT_USERNAME -p $MONGO_INITDB_ROOT_PASSWORD<<EOF
  use admin
  db.createUser(
          {
              user: "otelu",
              pwd: "otelp",
              roles: [
                  {
                      role: "clusterMonitor",
                      db: "admin"
                  }
              ]
          }
  );
EOF
}

echo "Configuring ${USER} permissions. . ."
end=$((SECONDS+20))
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
