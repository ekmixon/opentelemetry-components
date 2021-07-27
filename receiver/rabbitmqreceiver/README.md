# RabbitMQ Receiver

This receiver is fetches queue stats from a RabbitMQ instance using `/api/queues`. See [https://www.rabbitmq.com/monitoring.html](https://www.rabbitmq.com/monitoring.html) for more details.

Supported pipeline types: `metrics`

> :construction: This receiver is in **BETA**. Configuration fields and metric data model are subject to change.

## Prerequisites

Collecting metrics requires the ability to call `/api/queues`.  Please refer to [setup.sh](./testdata/scripts/setup.sh) for an example of how to configure these permissions. 

## Configuration

RabbitMQ receiver supports RabbitMQ version 3.8+

The following settings are required to create a database connection:
- `endpoint`
- `username`
- `password`

The following settings are optional:
- `collection_interval` (default = `10s`): This receiver collects metrics on an interval. This value must be a string readable by Golang's [time.ParseDuration](https://pkg.go.dev/time#ParseDuration). Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.

### Example Configuration

```yaml
receivers:
  rabbitmq:
    endpoint: localhost:15672
    username: otel
    password: $RABBITMQ_PASSWORD
    collection_interval: 10s
```

The full list of settings exposed for this receiver are documented [here](./config.go) with detailed sample configurations [here](./testdata/config.yaml).

## Metrics

Details about the metrics produced by this receiver can be found in [metadata.yaml](./metadata.yaml)
