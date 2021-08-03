# ElasticSearch Receiver

This receiver queries the ElasticSearch [statistics collector](https://www.elastic.co/guide/en/elasticsearch/reference/current/cluster-nodes-stats.html).

Supported pipeline types: `metrics`

> :construction: This receiver is in **BETA**. Configuration fields and metric data model are subject to change.

## Prerequisites

This receiver supports ElasticSearch versions 7.8+

## Configuration

The following settings are required to create a database connection:
- `endpoint`

The following settings are optional:
- `username`
- `password`
- `collection_interval` (default = `10s`): This receiver collects metrics on an interval. This value must be a string readable by Golang's [time.ParseDuration](https://pkg.go.dev/time#ParseDuration). Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`.

### Example Configuration

```yaml
receivers:
  elasticsearch:
    endpoint: localhost:9200
    username: otel
    password: password
    collection_interval: 10s
```

The full list of settings exposed for this receiver are documented [here](./config.go) with detailed sample configurations [here](./testdata/config.yaml).

## Metrics

Details about the metrics produced by this receiver can be found in [metadata.yaml](./metadata.yaml)