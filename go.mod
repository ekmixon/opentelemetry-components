module github.com/observiq/opentelemetry-components

go 1.16

require (
	github.com/client9/misspell v0.3.4
	github.com/golangci/golangci-lint v1.41.1
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudexporter v0.0.0-00010101000000-000000000000
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/observiqexporter v0.0.0-00010101000000-000000000000
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/collector v0.29.1-0.20210713183547-28091a238494
	go.opentelemetry.io/collector/model v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.18.1
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c
	golang.org/x/tools v0.1.5
)

replace go.opentelemetry.io/collector/model => go.opentelemetry.io/collector/model v0.0.0-20210713183547-28091a238494

replace github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudexporter => github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudexporter v0.29.1-0.20210714160339-8641edc22c7e

replace github.com/open-telemetry/opentelemetry-collector-contrib/exporter/observiqexporter => github.com/open-telemetry/opentelemetry-collector-contrib/exporter/observiqexporter v0.29.1-0.20210714160339-8641edc22c7e

replace github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor => github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.29.1-0.20210714160339-8641edc22c7e
