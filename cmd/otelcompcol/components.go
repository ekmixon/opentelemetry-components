package main

import (
	"github.com/GoogleCloudPlatform/opentelemetry-operations-collector/processor/normalizesumsprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/observiqexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/pprofextension"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/exporter/loggingexporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/extension/zpagesextension"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"

	"github.com/observiq/opentelemetry-components/receiver/couchdbreceiver"
	"github.com/observiq/opentelemetry-components/receiver/elasticsearchreceiver"
	"github.com/observiq/opentelemetry-components/receiver/httpdreceiver"
	"github.com/observiq/opentelemetry-components/receiver/mongodbreceiver"
	"github.com/observiq/opentelemetry-components/receiver/mysqlreceiver"
	"github.com/observiq/opentelemetry-components/receiver/postgresqlreceiver"
	"github.com/observiq/opentelemetry-components/receiver/rabbitmqreceiver"
)

// Get the factories for components we want to use.
// This includes all the defaults.
func components() (component.Factories, error) {
	var errs []error

	receivers, err := component.MakeReceiverFactoryMap(
		couchdbreceiver.NewFactory(),
		hostmetricsreceiver.NewFactory(),
		httpdreceiver.NewFactory(),
		otlpreceiver.NewFactory(),
		postgresqlreceiver.NewFactory(),
		mysqlreceiver.NewFactory(),
		mongodbreceiver.NewFactory(),
		rabbitmqreceiver.NewFactory(),
		elasticsearchreceiver.NewFactory(),
	)
	if err != nil {
		errs = append(errs, err)
	}

	processors, err := component.MakeProcessorFactoryMap(
		attributesprocessor.NewFactory(),
		filterprocessor.NewFactory(),
		resourceprocessor.NewFactory(),
		resourcedetectionprocessor.NewFactory(),
		metricstransformprocessor.NewFactory(),
		normalizesumsprocessor.NewFactory(),
	)
	if err != nil {
		errs = append(errs, err)
	}

	exporters, err := component.MakeExporterFactoryMap(
		fileexporter.NewFactory(),
		googlecloudexporter.NewFactory(),
		loggingexporter.NewFactory(),
		observiqexporter.NewFactory(),
		otlpexporter.NewFactory(),
		otlphttpexporter.NewFactory(),
	)
	if err != nil {
		errs = append(errs, err)
	}

	extensions, err := component.MakeExtensionFactoryMap(
		pprofextension.NewFactory(),
		zpagesextension.NewFactory(),
	)
	if err != nil {
		errs = append(errs, err)
	}

	factories := component.Factories{
		Receivers:  receivers,
		Processors: processors,
		Exporters:  exporters,
		Extensions: extensions,
	}

	return factories, consumererror.Combine(errs)

}
