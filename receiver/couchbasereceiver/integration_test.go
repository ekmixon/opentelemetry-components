package couchbasereceiver

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

	"github.com/observiq/opentelemetry-components/receiver/couchbasereceiver/internal/metadata"
)

func TestCouchbaseIntegration(t *testing.T) {
	// t.Run("Running docker version 6.6 on port 8092", func(t *testing.T) {
	// 	container := getContainer(t, containerRequest6_6)
	// 	defer func() {
	// 		require.NoError(t, container.Terminate(context.Background()))
	// 	}()
	// 	hostname, err := container.Host(context.Background())
	// 	require.NoError(t, err)

	// 	f := NewFactory()
	// 	cfg := f.CreateDefaultConfig().(*Config)
	// 	cfg.Endpoint = net.JoinHostPort(hostname, "8092")
	// 	cfg.Username = "otelu"
	// 	cfg.Password = "otelpassword"

	// 	consumer := new(consumertest.MetricsSink)
	// 	settings := componenttest.NewNopReceiverCreateSettings()
	// 	rcvr, err := f.CreateMetricsReceiver(context.Background(), settings, cfg, consumer)
	// 	require.NoError(t, err, "failed creating metrics receiver")
	// 	require.NoError(t, rcvr.Start(context.Background(), componenttest.NewNopHost()))
	// 	require.Eventuallyf(t, func() bool {
	// 		return len(consumer.AllMetrics()) > 0
	// 	}, 2*time.Minute, 1*time.Second, "failed to receive more than 0 metrics")

	// 	md := consumer.AllMetrics()[0]
	// 	require.Equal(t, 1, md.ResourceMetrics().Len())
	// 	ilms := md.ResourceMetrics().At(0).InstrumentationLibraryMetrics()
	// 	require.Equal(t, 1, ilms.Len())
	// 	metrics := ilms.At(0).Metrics()
	// 	require.NoError(t, rcvr.Shutdown(context.Background()))

	// 	validateIntegrationResult(t, metrics)
	// })

	t.Run("Running docker version 7.0 on port 8091", func(t *testing.T) {
		container := getContainer(t, containerRequest7_0)
		defer func() {
			require.NoError(t, container.Terminate(context.Background()))
		}()
		hostname, err := container.Host(context.Background())
		require.NoError(t, err)

		f := NewFactory()
		cfg := f.CreateDefaultConfig().(*Config)
		cfg.Endpoint = fmt.Sprintf("http://%s", net.JoinHostPort(hostname, "8091"))
		cfg.Username = "otelu"
		cfg.Password = "otelpassword"

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

		validateIntegrationResult(t, metrics)
	})
}

var (
	// containerRequest6_6 = testcontainers.ContainerRequest{
	// 	FromDockerfile: testcontainers.FromDockerfile{
	// 		Context:    path.Join(".", "testdata"),
	// 		Dockerfile: "Dockerfile.couchbase.6_6",
	// 	},
	// 	ExposedPorts: []string{"8091:8091"},
	// 	WaitingFor: wait.ForListeningPort("8091").
	// 		WithStartupTimeout(2 * time.Minute),
	// }
	containerRequest7_0 = testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.couchbase.7_0",
		},
		ExposedPorts: []string{"8091:8091"},
		WaitingFor: wait.ForListeningPort("8091").
			WithStartupTimeout(5 * time.Minute),
	}
)

type TestLogConsumer struct {
	t *testing.T
}

func (lc *TestLogConsumer) Accept(l testcontainers.Log) {
	lc.t.Log(string(l.Content))
}

func getContainer(t *testing.T, req testcontainers.ContainerRequest) testcontainers.Container {
	require.NoError(t, req.Validate())
	container, err := testcontainers.GenericContainer(
		context.Background(),
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
	require.NoError(t, err)

	logConsumer := TestLogConsumer{t}
	require.NoError(t, container.StartLogProducer(context.Background()))
	container.FollowOutput(&logConsumer)

	code, err := container.Exec(context.Background(), []string{"/setup.sh"})
	require.NoError(t, err)
	require.Equal(t, 0, code)

	require.NoError(t, container.StopLogProducer())
	return container
}

func validateIntegrationResult(t *testing.T, metric pdata.MetricSlice) {
	require.Equal(t, len(metadata.M.Names()), metric.Len())
}
