
# Mongodb Receiver

This receiver fetches stats from a Mongodb instance using the [golang
mongo driver](https://github.com/mongodb/mongo-go-driver). Stats are collected
via MongoDB's `dbStats` command.

Supported pipeline types: `metrics`

> :construction: This receiver is in **BETA**. Configuration fields and metric data model are subject to change.
## Prerequisites

This receiver supports MongoDB versions 4.0+

## Configuration

The following settings are required:
- `endpoint` (default: `localhost:27017`): The hostname/IP address and port of the mongodb instance

The following settings are optional:
- `user`: If authentication is required, the user can be provided here.
- `password`: If authentication is required, the password can be provided here.
- `collection_interval` (default = `10s`): This receiver collects metrics on an interval. This value must be a string readable by Golang's [time.ParseDuration](https://pkg.go.dev/time#ParseDuration). Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.

### Example Configuration

```yaml
receivers:
  mongodb:
    endpoint: localhost:27017
    user: otel
    password: $MONGODB_PASSWORD
    collection_interval: 10s
```

The full list of settings exposed for this receiver are documented [here](./config.go) with detailed sample configurations [here](./testdata/config.yaml).

## Metrics

Details about the metrics produced by this receiver can be found in [metadata.yaml](./metadata.yaml)
