package normalizesumsprocessor

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configtest"
)

func TestLoadConfig(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.NoError(t, err)

	factory := NewFactory()
	factories.Processors[typeStr] = factory

	cfg, err := configtest.LoadConfigAndValidate(path.Join(".", "testdata", "transform_all_config.yaml"), factories)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	id := config.NewID(typeStr)
	settings := config.NewProcessorSettings(id)
	p1 := cfg.Processors[id]
	expectedCfg := &Config{
		ProcessorSettings: &settings,
	}
	assert.Equal(t, p1, expectedCfg)
}
