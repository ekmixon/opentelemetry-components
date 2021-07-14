package normalizesumsprocessor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configcheck"
	"go.opentelemetry.io/collector/consumer/consumertest"
)

func TestCreateDefaultConfig(t *testing.T) {
	cfg := createDefaultConfig()
	assert.NoError(t, configcheck.ValidateConfig(cfg))
	assert.NotNil(t, cfg)
}

func TestCreateProcessor(t *testing.T) {
	mp, err := createMetricsProcessor(context.Background(), component.ProcessorCreateSettings{}, createDefaultConfig(), consumertest.NewNop())
	assert.NoError(t, err)
	assert.NotNil(t, mp)
}
