package normalizesumsprocessor

import "go.opentelemetry.io/collector/config"

// Config defines configuration for Resource processor.
type Config struct {
	*config.ProcessorSettings `mapstructure:"-"`
}
