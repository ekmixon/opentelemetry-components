# RabbitMQ Receiver

The RabbitMQ receiver is designed to retrieve RabbitMQ queue stats from a single RabbitMQ instance and build metrics from that data. Then send them to the next consumer at a configurable interval. Configuration options for connecting to the database are also required.

> :construction: This receiver is in beta and configuration fields are subject to change.

## Details 

RabbitMQ stats are pulled using `/api/queues` which returns an array of queues with the applicable metrics for each queue (see [https://www.rabbitmq.com/monitoring.html](https://www.rabbitmq.com/monitoring.html) for details). The RabbitMQ receiver extracts values from the results and converts them to open telemetry metrics. Details about the metrics produce by the RabbitMQ receiver can be found in [metadata.yaml](metadata.yaml).

## Prerequisites

Collecting metrics requires the ability to call `/api/queues`.  Please refer to [setup.sh](./testdata/scripts/setup.sh) for an example of how to configure these permissions. 

## Configuration

RabbitMQ receiver supports RabbitMQ version 3.8 and above.

The following settings are required to create a database connection:

- `username`
- `password`
- `endpoint`

The following settings are optional:

- `collection_interval` (default = `10s`): This receiver runs on an interval.
Each time it runs, it queries rabbitmq, creates metrics, and sends them to the
next consumer. The `collection_interval` configuration option tells this
receiver the duration between runs. This value must be a string readable by
Golang's `ParseDuration` function (example: `1h30m`). Valid time units are
`ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.

Example:

```yaml
receivers:
  rabbitmq:
    collection_interval: 10s
    username: otel
    password: otel
    endpoint: localhost:15672
```

The full list of settings exposed for this receiver are documented [here](./config.go)
with detailed sample configurations [here](./testdata/config.yaml).


