// +build integration

package rabbitmqreceiver

import (
	"context"
	"fmt"
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

	"github.com/observiq/opentelemetry-components/receiver/rabbitmqreceiver/internal/metadata"
)

func TestRabbitMQScraperHappyPath(t *testing.T) {
	container := getContainer(t, containerRequest3_8)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = fmt.Sprintf("http://%s", net.JoinHostPort(hostname, "15672"))
	cfg.Password = "dev"
	cfg.Username = "dev"

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
	require.NoError(t, rcvr.Shutdown(context.Background()))

	validateResult(t, metrics)
}

var (
	containerRequest3_8 = testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.rabbitmq",
		},
		ExposedPorts: []string{"15672:15672"},
		WaitingFor: wait.ForListeningPort("15672").
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

	code, err := container.Exec(context.Background(), []string{"/setup.sh"})
	require.NoError(t, err)
	require.Equal(t, 0, code)
	time.Sleep(time.Second * 6) // TODO customize wait.Strategy
	return container
}

func validateResult(t *testing.T, metrics pdata.MetricSlice) {
	require.Equal(t, len(metadata.M.Names()), metrics.Len())

	// TODO thorough validation
}
