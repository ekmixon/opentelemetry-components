# CouchDB Receiver

This receiver fetches stats from a couchdb server using the `/_node/{node-name}/_stats/couchdb` [endpoint](https://docs.couchdb.org/en/latest/api/server/common.html#node-node-name-stats).

Supported pipeline types: `metrics`

> :construction: This receiver is in **BETA**. Configuration fields and metric data model are subject to change.

## Prerequisites

This receiver supports Couchdb versions `3.0` and `3.1`.

## Configuration

The following settings are required:
- `endpoint`
- `username`
- `password`
- `nodename` (default: _local)

The following settings are optional:
- `collection_interval` (default = `10s`): This receiver collects metrics on an interval. This value must be a string readable by Golang's [time.ParseDuration](https://pkg.go.dev/time#ParseDuration). Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.

### Example Configuration

```yaml
receivers:
  couchdb:
    endpoint: localhost:5984
    username: otelu
    password: $couchdb_PASSWORD
    nodename: _local
    collection_interval: 10s
```

The full list of settings exposed for this receiver are documented [here](./config.go) with detailed sample configurations [here](./testdata/config.yaml).

## Metrics

Details about the metrics produced by this receiver can be found in [metadata.yaml](./metadata.yaml)
