// +build integration

package mongodbreceiver

import (
	"context"
	"net"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/model/pdata"

	"github.com/observiq/opentelemetry-components/receiver/mongodbreceiver/internal/metadata"
)

func TestMongoDBIntegration(t *testing.T) {
	container := getContainer(t, containerRequest4_0)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = net.JoinHostPort(hostname, "37017")
	cfg.Username = "otel"
	cfg.Password = "otel"

	consumer := new(consumertest.MetricsSink)
	settings := componenttest.NewNopReceiverCreateSettings()
	rcvr, err := f.CreateMetricsReceiver(context.Background(), settings, cfg, consumer)
	require.NoError(t, err, "failed creating metrics receiver")
	require.NoError(t, rcvr.Start(context.Background(), componenttest.NewNopHost()))
	require.Eventuallyf(t, func() bool {
		return len(consumer.AllMetrics()) > 0
	}, 2*time.Minute, 1*time.Second, "failed to receive more than 0 metrics")

	md := consumer.AllMetrics()[0]
	require.Equal(t, 1, md.ResourceMetrics().Len())
	ilms := md.ResourceMetrics().At(0).InstrumentationLibraryMetrics()
	require.Equal(t, 1, ilms.Len())
	metrics := ilms.At(0).Metrics()
	require.Equal(t, 13, metrics.Len())
	require.NoError(t, rcvr.Shutdown(context.Background()))

	validateResult(t, metrics)
}

var (
	containerRequest4_0 = testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.mongodb",
		},
		ExposedPorts: []string{"37017:27017"},
		WaitingFor: wait.ForListeningPort("27017").
			WithStartupTimeout(2 * time.Minute),
	}
)

func getContainer(t *testing.T, req testcontainers.ContainerRequest) testcontainers.Container {
	require.NoError(t, req.Validate())
	container, err := testcontainers.GenericContainer(
		context.Background(),
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
	require.NoError(t, err)
	return container
}

func validateResult(t *testing.T, metrics pdata.MetricSlice) {
	names := metadata.Metrics.Names()
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
