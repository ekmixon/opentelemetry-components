package postgresqlreceiver

import (
	"context"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/testbed/testbed"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

func TestType(t *testing.T) {
	factory := NewFactory()
	ft := factory.Type()
	require.EqualValues(t, "postgresql", ft)
}

func TestValidConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.Username = "otel"
	cfg.Password = "otel"
	cfg.Endpoint = "localhost:5432"
	require.NoError(t, cfg.Validate())
}

func TestCreateMetricsReceiver(t *testing.T) {
	factory := NewFactory()
	metricsReceiver, err := factory.CreateMetricsReceiver(
		context.Background(),
		component.ReceiverCreateSettings{},
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
