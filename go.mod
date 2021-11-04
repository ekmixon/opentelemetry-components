module github.com/observiq/opentelemetry-components

go 1.16

require (
	github.com/GoogleCloudPlatform/opentelemetry-operations-collector v0.0.3-0.20211006225201-9fda76b49539
	github.com/go-sql-driver/mysql v1.6.0
	github.com/lib/pq v1.10.3
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter v0.36.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudexporter v0.36.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/observiqexporter v0.36.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/pprofextension v0.36.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.36.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.36.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor v0.36.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.36.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.36.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver v0.36.0
	github.com/stretchr/testify v1.7.0
	github.com/testcontainers/testcontainers-go v0.11.1
	go.mongodb.org/mongo-driver v1.7.2
	go.opentelemetry.io/collector v0.36.0
	go.opentelemetry.io/collector/model v0.36.0
	go.uber.org/multierr v1.7.0
	go.uber.org/zap v1.19.1
	golang.org/x/sys v0.0.0-20210908233432-aa78b53d3365
)
