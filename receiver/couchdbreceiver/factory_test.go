package couchdbreceiver

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/multierr"
)

func TestType(t *testing.T) {
	factory := NewFactory()
	ft := factory.Type()
	require.EqualValues(t, "couchdb", ft)
}

func TestValidConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.Username = "otelu"
	cfg.Password = "otelp"
	cfg.Endpoint = "localhost:5984"
	require.NoError(t, cfg.Validate())
}

func TestCreateMetricsReceiver(t *testing.T) {
	t.Run("validation error, missing config fields", func(t *testing.T) {
		factory := NewFactory()
		metricsReceiver, err := factory.CreateMetricsReceiver(
			context.Background(),
			component.ReceiverCreateSettings{},
			&Config{
				ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
					ReceiverSettings:   config.NewReceiverSettings(config.NewID("couchdb")),
					CollectionInterval: 10 * time.Second,
				},
			},
			consumertest.NewNop(),
		)
		require.Error(t, err)
		expectedErr := multierr.Combine(
			errors.New(ErrNoUsername),
			errors.New(ErrNoPassword),
			errors.New(ErrNoEndpoint),
		)
		require.Equal(t, expectedErr, err)
		require.Nil(t, metricsReceiver)
	})

	t.Run("no error", func(t *testing.T) {
		factory := NewFactory()
		metricsReceiver, err := factory.CreateMetricsReceiver(
			context.Background(),
			component.ReceiverCreateSettings{},
			&Config{
				ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
					ReceiverSettings:   config.NewReceiverSettings(config.NewID("couchdb")),
					CollectionInterval: 10 * time.Second,
				},
				Username: "otelu",
				Password: "otelp",
				Endpoint: "localhost:5984",
			},
			consumertest.NewNop(),
		)
		require.NoError(t, err)
		require.NotNil(t, metricsReceiver)
	})
}
