package httpdreceiver

import (
	"context"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/testbed/testbed"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
)

func TestType(t *testing.T) {
	factory := NewFactory()
	ft := factory.Type()
	require.EqualValues(t, "httpd", ft)
}

func TestValidConfig(t *testing.T) {
	factory := NewFactory()
	err := factory.CreateDefaultConfig().Validate()
	require.NoError(t, err)
}

func TestCreateMetricsReceiver(t *testing.T) {
	factory := NewFactory()
	metricsReceiver, err := factory.CreateMetricsReceiver(
		context.Background(),
		component.ReceiverCreateSettings{},
		&Config{
			ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
				CollectionInterval: 10 * time.Second,
			},
		},
		&testbed.MockMetricConsumer{},
	)
	require.NoError(t, err)
	require.NotNil(t, metricsReceiver)
}
