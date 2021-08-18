package postgresqlreceiver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configcheck"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.opentelemetry.io/collector/testbed/testbed"
	"go.uber.org/zap"
)

func TestType(t *testing.T) {
	factory := NewFactory()
	ft := factory.Type()
	require.EqualValues(t, "postgresql", ft)
}

func TestValidConfig(t *testing.T) {
	factory := NewFactory()
	err := configcheck.ValidateConfig(factory.CreateDefaultConfig())
	require.NoError(t, err)
}

func TestCreateMetricsReceiver(t *testing.T) {
	factory := NewFactory()
	metricsReceiver, err := factory.CreateMetricsReceiver(
		context.Background(),
		component.ReceiverCreateSettings{Logger: zap.NewNop()},
		&Config{
			ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
				ReceiverSettings:   config.NewReceiverSettings(config.NewID("postgresql")),
				CollectionInterval: 10 * time.Second,
			},
			Username: "otel",
			Password: "otel",
			Database: "otel",
			Endpoint: "localhost:5432",
		},
		&testbed.MockMetricConsumer{},
	)
	require.NoError(t, err)
	require.NotNil(t, metricsReceiver)

}
