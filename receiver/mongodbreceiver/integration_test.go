// +build integration

package mongodbreceiver

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/model/pdata"

	"github.com/observiq/opentelemetry-components/receiver/mongodbreceiver/internal/metadata"
)

func mongodbContainer(t *testing.T) testcontainers.Container {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.mongodb",
		},
		ExposedPorts: []string{"37017:27017"},
		WaitingFor:   wait.ForListeningPort("27017"),
	}

	require.NoError(t, req.Validate())

	mongodb, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	return mongodb
}

func TestIntegration(t *testing.T) {
	container := mongodbContainer(t)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = net.JoinHostPort(hostname, "37017")

	user := "otel"
	pass := "otel"
	cfg.User = &user
	cfg.Password = &pass

	consumer := new(consumertest.MetricsSink)

	rcvr, err := f.CreateMetricsReceiver(context.Background(), componenttest.NewNopReceiverCreateSettings(), cfg, consumer)
	require.NoError(t, err, "failed creating metrics receiver")
	require.NoError(t, rcvr.Start(context.Background(), componenttest.NewNopHost()))

	require.Eventuallyf(t, func() bool {
		return len(consumer.AllMetrics()) > 0
	}, 15*time.Second, 1*time.Second, "failed to receive more than 0 metrics")

	md := consumer.AllMetrics()[0]

	require.Equal(t, 1, md.ResourceMetrics().Len())

	ilms := md.ResourceMetrics().At(0).InstrumentationLibraryMetrics()
	require.Equal(t, 1, ilms.Len())

	metrics := ilms.At(0).Metrics()
	require.Equal(t, 13, metrics.Len())

	assertAllMetricNamesArePresent(t, metadata.Metrics.Names(), metrics)

	assert.NoError(t, rcvr.Shutdown(context.Background()))
}

func assertAllMetricNamesArePresent(t *testing.T, names []string, metrics pdata.MetricSlice) {
	seen := make(map[string]bool, len(names))
	for i := range names {
		seen[names[i]] = false
	}

	for i := 0; i < metrics.Len(); i++ {
		seen[metrics.At(i).Name()] = true
	}

	for k, v := range seen {
		if !v {
			t.Fatalf("Did not find metric %q", k)
		}
	}
}
