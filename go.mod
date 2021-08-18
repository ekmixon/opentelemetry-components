module github.com/observiq/opentelemetry-components

go 1.16

require (
	github.com/fatih/color v1.12.0 // indirect
	github.com/go-sql-driver/mysql v1.6.0
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/lib/pq v1.2.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudexporter v0.30.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/observiqexporter v0.30.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.30.0
	github.com/stretchr/testify v1.7.0
	github.com/testcontainers/testcontainers-go v0.11.1
	go.mongodb.org/mongo-driver v1.7.1
	go.opentelemetry.io/collector v0.31.0
	go.opentelemetry.io/collector/model v0.31.0
	go.uber.org/multierr v1.7.0
	go.uber.org/zap v1.18.1
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c
	golang.org/x/tools v0.1.5 // indirect
)
