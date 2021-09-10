# CouchBase Receiver

This receiver fetches stats from a couchbase server using the `/pools/default` [Cluster endpoint](https://docs.couchbase.com/server/current/rest-api/rest-cluster-intro.html) and `/pools/default/buckets` [Buckets endpoint](https://docs.couchbase.com/server/current/rest-api/rest-bucket-intro.html).

Supported pipeline types: `metrics`

> :construction: This receiver is in **BETA**. Configuration fields and metric data model are subject to change.

## Prerequisites

This receiver supports Couchbase versions `6.6` and `7.0`.

## Configuration

The following settings are required:
- `endpoint`
- `username`
- `password`

The following settings are optional:
- `collection_interval` (default = `10s`): This receiver collects metrics on an interval. This value must be a string readable by Golang's [time.ParseDuration](https://pkg.go.dev/time#ParseDuration). Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.

### Example Configuration

```yaml
receivers:
  couchbase:
    endpoint: localhost:8091
    username: otelu
    password: $couchbase_PASSWORD
    collection_interval: 60s
```

The full list of settings exposed for this receiver are documented [here](./config.go) with detailed sample configurations [here](./testdata/config.yaml).

## Metrics

Details about the metrics produced by this receiver can be found in [metadata.yaml](./metadata.yaml)
