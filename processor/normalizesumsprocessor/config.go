package normalizesumsprocessor

import "go.opentelemetry.io/collector/config"

// Config defines configuration for Resource processor.
type Config struct {
	*config.ProcessorSettings `mapstructure:"-"`

	// Transforms describes the metrics that will be transformed
	Transforms []Transform `mapstructure:"transforms"`
}

// Transform defines the transformation applied to the specific metric
type Transform struct {
	// MetricName is used to select the metric to operate on.
	// REQUIRED
	MetricName string `mapstructure:"metric_name"`

	// NewName specifies the name of the new metric after transforming.
	NewName string `mapstructure:"new_name"`
}
