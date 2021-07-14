package main

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/observiqexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/exporter/fileexporter"
	"go.opentelemetry.io/collector/exporter/loggingexporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/extension/pprofextension"
	"go.opentelemetry.io/collector/extension/zpagesextension"
	"go.opentelemetry.io/collector/processor/attributesprocessor"
	"go.opentelemetry.io/collector/processor/filterprocessor"
	"go.opentelemetry.io/collector/processor/resourceprocessor"
	"go.opentelemetry.io/collector/receiver/hostmetricsreceiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"

	"github.com/observiq/opentelemetry-components/processor/normalizesumsprocessor"
	"github.com/observiq/opentelemetry-components/receiver/httpdreceiver"
)

// Get the factories for components we want to use.
// This includes all the defaults.
func components() (component.Factories, error) {
	var errs []error

	receivers, err := component.MakeReceiverFactoryMap(
		hostmetricsreceiver.NewFactory(),
		httpdreceiver.NewFactory(),
		otlpreceiver.NewFactory(),
	)
	if err != nil {
		errs = append(errs, err)
	}

	processors, err := component.MakeProcessorFactoryMap(
		attributesprocessor.NewFactory(),
		filterprocessor.NewFactory(),
		normalizesumsprocessor.NewFactory(),
		resourceprocessor.NewFactory(),
		resourcedetectionprocessor.NewFactory(),
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
