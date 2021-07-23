# PostgreSQL Receiver

The PostgreSQL receiver is designed to query a single PostgreSQL database instance, build metrics from that data, and send them to the next consumer at a configurable interval. Configuration options for connecting to the database are also required.

> :construction: This receiver is in beta and configuration fields are subject to change.

## Details

The PostgreSQL receiver uses PostgreSQL's statistics collector to collect information about server activity (see [https://www.postgresql.org/docs/9.6/monitoring-stats.html](https://www.postgresql.org/docs/9.6/monitoring-stats.html) for details). The PostgreSQL receiver extracts values from the statistics collector and converts them to open telemetry metrics. Details about the metrics produce by the PostgreSQL receiver can be found in [metadata.yaml](metadata.yaml).

## Configuration

PostgreSQL receiver supports PostgreSQL version 9.6

The following settings are required to create a database connection:

- `username`
- `password`
- `database`
- `endpoint`


- `collection_interval` (default = `10s`): This receiver runs on an interval.
Each time it runs, it queries postgresql, creates metrics, and sends them to the
next consumer. The `collection_interval` configuration option tells this
receiver the duration between runs. This value must be a string readable by
Golang's `ParseDuration` function (example: `1h30m`). Valid time units are
`ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.

Example:

```yaml
receivers:
  postgresql:
    collection_interval: 10s
    username: otel
    password: otel
    database: otel
    endpoint: 127.0.0.1:5432
```

The full list of settings exposed for this receiver are documented [here](./config.go)
with detailed sample configurations [here](./testdata/config.yaml).


