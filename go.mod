module github.com/observiq/opentelemetry-components

go 1.16

require (
	github.com/client9/misspell v0.3.4
	github.com/golangci/golangci-lint v1.41.1
	github.com/lib/pq v1.9.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudexporter v0.30.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/observiqexporter v0.30.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.30.0
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.30.0
	go.opentelemetry.io/collector/model v0.30.0
	go.uber.org/zap v1.18.1
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c
	golang.org/x/tools v0.1.5
)
