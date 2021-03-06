package postgresqlreceiver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer/consumertest"
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
	cfg.Host = "localhost"
	cfg.Port = 5432
	require.NoError(t, cfg.Validate())
}

func TestCreateMetricsReceiver(t *testing.T) {
	factory := NewFactory()
	metricsReceiver, err := factory.CreateMetricsReceiver(
		context.Background(),
		componenttest.NewNopReceiverCreateSettings(),
		&Config{
			ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
				ReceiverSettings:   config.NewReceiverSettings(config.NewComponentID("postgresql")),
				CollectionInterval: 10 * time.Second,
			},
			Username:  "otel",
			Password:  "otel",
			Databases: []string{"otel"},
			Host:      "localhost",
			Port:      5432,
		},
		consumertest.NewNop(),
	)
	require.NoError(t, err)
	require.NotNil(t, metricsReceiver)

}
