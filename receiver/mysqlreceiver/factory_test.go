package mysqlreceiver

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
	require.EqualValues(t, "mysql", ft)
}

func TestValidConfig(t *testing.T) {
	factory := NewFactory()
	err := configcheck.ValidateConfig(factory.CreateDefaultConfig())
	require.NoError(t, err)
}

func TestCreateMetricsReceiver(t *testing.T) {
	testCases := []struct {
		desc        string
		username    string
		password    string
		endpoint    string
		expectedErr string
	}{
		{
			desc:        "missing username",
			username:    "",
			password:    "otel",
			endpoint:    "localhost:3306",
			expectedErr: ErrNoUsername,
		},
		{
			desc:        "missing password",
			username:    "otel",
			password:    "",
			endpoint:    "localhost:3306",
			expectedErr: ErrNoPassword,
		},
		{
			desc:        "missing endpoint",
			username:    "otel",
			password:    "otel",
			endpoint:    "",
			expectedErr: ErrNoEndpoint,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			factory := NewFactory()
			metricsReceiver, err := factory.CreateMetricsReceiver(
				context.Background(),
				component.ReceiverCreateSettings{Logger: zap.NewNop()},
				&Config{
					ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
						ReceiverSettings:   config.NewReceiverSettings(config.NewID("mysql")),
						CollectionInterval: 10 * time.Second,
					},
					Username: tC.username,
					Password: tC.password,
					Endpoint: tC.endpoint,
				},
				&testbed.MockMetricConsumer{},
			)
			require.Error(t, err)
			require.EqualError(t, err, tC.expectedErr)
			require.Nil(t, metricsReceiver)
		})
	}

	t.Run("happy path", func(t *testing.T) {
		factory := NewFactory()
		metricsReceiver, err := factory.CreateMetricsReceiver(
			context.Background(),
			component.ReceiverCreateSettings{Logger: zap.NewNop()},
			&Config{
				ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
					ReceiverSettings:   config.NewReceiverSettings(config.NewID("mysql")),
					CollectionInterval: 10 * time.Second,
				},
				Username: "otel",
				Password: "otel",
				Endpoint: "localhost:3306",
			},
			&testbed.MockMetricConsumer{},
		)
		require.NoError(t, err)
		require.NotNil(t, metricsReceiver)
	})
}
