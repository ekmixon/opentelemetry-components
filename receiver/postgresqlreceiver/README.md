# PostgreSQL Receiver

This receiver queries PostgreSQL's statistics collector [https://www.postgresql.org/docs/9.6/monitoring-stats.html](https://www.postgresql.org/docs/9.6/monitoring-stats.html).

Supported pipeline types: `metrics`

> :construction: This receiver is in **BETA**. Configuration fields and metric data model are subject to change.

## Prerequisites

This receiver supports PostgreSQL versions 9.6+

## Configuration

The following settings are required to create a database connection:
- `endpoint`
- `username`
- `password`
- `database`

The following settings are optional:
- `collection_interval` (default = `10s`): This receiver collects metrics on an interval. This value must be a string readable by Golang's [time.ParseDuration](https://pkg.go.dev/time#ParseDuration). Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.

### Example Configuration

```yaml
receivers:
  postgresql:
    endpoint: 127.0.0.1:5432
    username: otel
    password: $POSTGRESQL_PASSWORD
    database: otel
    collection_interval: 10s
```

The full list of settings exposed for this receiver are documented [here](./config.go) with detailed sample configurations [here](./testdata/config.yaml).

## Metrics

Details about the metrics produced by this receiver can be found in [metadata.yaml](./metadata.yaml)
