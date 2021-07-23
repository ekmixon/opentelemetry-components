# MySQL Receiver

The MySQL receiver is designed to retrieve MySQL Global Status and InnoDB data from a single MySQL instance, build, build metrics from that data, and send them to the next consumer at a configurable interval. Configuration options for connecting to the database are also required.

> :construction: This receiver is in beta and configuration fields are subject to change.

## Details

The MySQL `SHOW GLOBAL STATUS` and `information_schema.innodb_metrics` table contain information and statistics about a MySQL server status (see [https://dev.mysql.com/doc/refman/8.0/en/server-status-variables.html](https://dev.mysql.com/doc/refman/8.0/en/server-status-variables.html) and for details). The MySQL receiver extracts values from the results and converts them to open telemetry metrics. Details about the metrics produce by the MySQL receiver can be found in [metadata.yaml](metadata.yaml).

## Prerequisites

Collecting most metrics requires the ability to execute `SHOW GLOBAL STATUS`. The `buffer_pool_size` metric requires access to the `information_schema.innodb_metrics` table. Please refer to [setup.sh](/receiver/mysqlreceiver/testdata/scripts/setup.sh) for an example of how to configure these permissions. 

## Configuration

> :information_source: This receiver is in beta and configuration fields are subject to change.

MySQL receiver supports MySQL version 8.0

The following settings are required to create a database connection:

- `username`
- `password`
- `endpoint`

The following settings are optional:

- `database`: The database name.

- `collection_interval` (default = `10s`): This receiver runs on an interval.
Each time it runs, it queries mysql, creates metrics, and sends them to the
next consumer. The `collection_interval` configuration option tells this
receiver the duration between runs. This value must be a string readable by
Golang's `ParseDuration` function (example: `1h30m`). Valid time units are
`ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.

Example:

```yaml
receivers:
  mysql:
    collection_interval: 10s
    username: otel
    password: otel
    database: otel
    endpoint: localhost:3306
```

The full list of settings exposed for this receiver are documented [here](./config.go)
with detailed sample configurations [here](./testdata/config.yaml).


