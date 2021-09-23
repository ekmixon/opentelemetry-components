package normalizesumsprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

const (
	// The value of "type" key in configuration.
	typeStr = "normalizesums"
)

func NewFactory() component.ProcessorFactory {
	return processorhelper.NewFactory(
		typeStr,
		createDefaultConfig,
		processorhelper.WithMetrics(createMetricsProcessor))
}

func createDefaultConfig() config.Processor {
	settings := config.NewProcessorSettings(config.NewID(typeStr))
	return &Config{
		ProcessorSettings: &settings,
	}
}

var processorCapabilities = consumer.Capabilities{MutatesData: true}

func createMetricsProcessor(
	_ context.Context,
	params component.ProcessorCreateSettings,
	cfg config.Processor,
	nextConsumer consumer.Metrics,
) (component.MetricsProcessor, error) {
	oCfg := cfg.(*Config)
	if err := validateConfiguration(oCfg); err != nil {
		return nil, err
	}
	metricsProcessor := newNormalizeSumsProcessor(params.Logger)
	return processorhelper.NewMetricsProcessor(
		cfg,
		nextConsumer,
		metricsProcessor.ProcessMetrics,
		processorhelper.WithCapabilities(processorCapabilities))
}

// validateConfiguration validates the input configuration has all of the required fields for the processor
// An error is returned if there are any invalid inputs.
func validateConfiguration(config *Config) error {
	return nil
}
