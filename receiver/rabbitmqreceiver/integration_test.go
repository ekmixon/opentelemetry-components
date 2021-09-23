//go:build integration
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
	exists := make(map[string]bool)

	unenumLabelSet := []string{
		metadata.L.Queue,
	}

	enumLabelSet := []string{
		metadata.L.State,
	}

	for i := 0; i < metrics.Len(); i++ {
		m := metrics.At(i)
		require.Contains(t, metadata.M.Names(), m.Name())

		metricIntr := metadata.M.ByName(m.Name())
		require.Equal(t, metricIntr.New().DataType(), m.DataType())
		var dps pdata.NumberDataPointSlice
		switch m.DataType() {
		case pdata.MetricDataTypeGauge:
			dps = m.Gauge().DataPoints()
		case pdata.MetricDataTypeSum:
			dps = m.Sum().DataPoints()
		}

		for j := 0; j < dps.Len(); j++ {
			key := m.Name()
			dp := dps.At(j)

			for _, label := range unenumLabelSet {
				_, ok := dp.Attributes().Get(label)
				if ok {
					key = key + " " + label
				}
			}

			for _, label := range enumLabelSet {
				labelVal, ok := dp.Attributes().Get(label)
				if ok {
					key += " " + labelVal.AsString()
				}
			}
			exists[key] = true
		}
	}

	// TODO: uncomment These when load gen is added
	require.Equal(t, map[string]bool{
		"rabbitmq.consumers queue": true,
		// "rabbitmq.delivery_rate queue":               true,
		// "rabbitmq.publish_rate queue":                true,
		"rabbitmq.num_messages queue unacknowledged": true,
		"rabbitmq.num_messages queue ready":          true,
		"rabbitmq.num_messages queue total":          true,
	}, exists)
}
