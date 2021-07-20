# Mongodb Receiver

This receiver can fetch stats from a Mongodb instance using the [golang
mongo driver](https://github.com/mongodb/mongo-go-driver). Stats are collected
using the dbStats command.

> :construction: This receiver is currently in **BETA**.

## Details

## Configuration

> :information_source: This receiver is in beta and configuration fields are subject to change.

The following settings are required:

- `endpoint` (default: `localhost:27017`): The hostname/IP address and port of the mongodb instance

The following settings are optional:

- `user`: If authentication is required, the user can be provided here.
- `password`: If authentication is required, the password can be provided here.
- `collection_interval` (default = `10s`): This receiver runs on an interval.
Each time it runs, it queries mongodb, creates metrics, and sends them to the
next consumer. The `collection_interval` configuration option tells this
receiver the duration between runs. This value must be a string readable by
Golang's `ParseDuration` function (example: `1h30m`). Valid time units are
`ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.

Example:

```yaml
receivers:
  mongodb:
    endpoint: "localhost:27017"
    user: otel
    password: otel
    collection_interval: 60s
```

The full list of settings exposed for this receiver are documented [here](./config.go)
with detailed sample configurations [here](./testdata/config.yaml).
