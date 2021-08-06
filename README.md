## Purpose

This repository hosts components that may be imported into distributions of the [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector).

## Unique Components

The following components are based in this repository:

Receivers:
- [couchdb](/receiver/couchdbreceiver/)
- [elasticsearch](/receiver/elasticsearchreceiver/)
- [httpd](/receiver/httpdreceiver/)
- [mysql](/receiver/mysqlreceiver/)
- [mongodb](/receiver/mongodbreceiver/)
- [rabbitmq](/receiver/rabbitmqreceiver/)
- [postgresql](/receiver/postgresqlreceiver/)

Processors:
- [normalizesums](/processor/normalizesumsprocessor/)


## Upstream Components

Several components are included from the OpenTelemetry Collector and its Contrib variant. These are primarily included to facilitate development and testing of new components:

Receivers:
- [hostmetrics](https://github.com/open-telemetry/opentelemetry-collector/tree/main/receiver/hostmetricsreceiver)
- [otlp](https://github.com/open-telemetry/opentelemetry-collector/tree/main/receiver/otlpreceiver)

Processors:
- [attributes](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/attributesprocessor)
- [filter](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/filterprocessor)
- [resource](https://github.com/open-telemetry/opentelemetry-collector/tree/main/processor/resourceprocessor)
- [resourcedetectionprocessor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/resourcedetectionprocessor)

Exporters:
- [file](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/fileexporter)
- [googlecloud](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/googlecloudexporter)
- [logging](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/loggingexporter)
- [observiq](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/observiqexporter)
- [otlp](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlpexporter)
- [otlphttp](https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter)

Extensions:
- [pprof](https://github.com/open-telemetry/opentelemetry-collector/tree/main/extension/pprofextension)
- [zpages](https://github.com/open-telemetry/opentelemetry-collector/tree/main/extension/zpagesextension)


## Development

This project is primarily focused on the development and hosting of new components. As a matter of developer convenience, the project itself is structured as a distribution of the collector. This means that it can easily be built into an executable that can run in the same manner as the collector. 

### Initial Setup

Clone this repository, and run `make install-tools`.

### Tooling

Typical functionality is included, such as `make build` and `make test`. Check out the [Makefile](./Makefile) for additional tooling.
