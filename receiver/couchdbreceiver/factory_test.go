package couchdbreceiver

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configcheck"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.opentelemetry.io/collector/testbed/testbed"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

func TestType(t *testing.T) {
	factory := NewFactory()
	ft := factory.Type()
	require.EqualValues(t, "couchdb", ft)
}

func TestValidConfig(t *testing.T) {
	factory := NewFactory()
	err := configcheck.ValidateConfig(factory.CreateDefaultConfig())
	require.NoError(t, err)
}

func TestCreateMetricsReceiver(t *testing.T) {
	t.Run("validation error, missing config fields", func(t *testing.T) {
		factory := NewFactory()
		metricsReceiver, err := factory.CreateMetricsReceiver(
			context.Background(),
			component.ReceiverCreateSettings{Logger: zap.NewNop()},
			&Config{
				ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
					ReceiverSettings:   config.NewReceiverSettings(config.NewID("couchdb")),
					CollectionInterval: 10 * time.Second,
				},
			},
			&testbed.MockMetricConsumer{},
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
			component.ReceiverCreateSettings{Logger: zap.NewNop()},
			&Config{
				ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
					ReceiverSettings:   config.NewReceiverSettings(config.NewID("couchdb")),
					CollectionInterval: 10 * time.Second,
				},
				Username: "otelu",
				Password: "otelp",
				Endpoint: "localhost:5984",
			},
			&testbed.MockMetricConsumer{},
		)
		require.NoError(t, err)
		require.NotNil(t, metricsReceiver)
	})
}
