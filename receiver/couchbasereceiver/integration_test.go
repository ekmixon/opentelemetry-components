//go:build integration
// +build integration

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
	t.Run("Running docker version 6.6", func(t *testing.T) {
		t.Parallel()
		container := getContainer(t, containerRequest6_6)
		defer func() {
			require.NoError(t, container.Terminate(context.Background()))
		}()
		hostname, err := container.Host(context.Background())
		require.NoError(t, err)

		f := NewFactory()
		cfg := f.CreateDefaultConfig().(*Config)
		cfg.Endpoint = fmt.Sprintf("http://%s", net.JoinHostPort(hostname, "8092"))
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

	t.Run("Running docker version 7.0", func(t *testing.T) {
		t.Parallel()
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
	containerRequest6_6 = testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.couchbase.6_6",
		},
		ExposedPorts: []string{"8092:8091"},
		WaitingFor: wait.ForListeningPort("8091").
			WithStartupTimeout(2 * time.Minute),
	}
	containerRequest7_0 = testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.couchbase.7_0",
		},
		ExposedPorts: []string{"8091:8091"},
		WaitingFor: wait.ForListeningPort("8091").
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
	return container
}

func validateIntegrationResult(t *testing.T, metric pdata.MetricSlice) {
	require.Equal(t, len(metadata.M.Names()), metric.Len())
	for i := 0; i < metric.Len(); i++ {
		m := metric.At(i)
		switch m.Name() {
		case metadata.M.CouchbaseBDataUsed.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 2, dps.Len())

			metricsMap := map[string]bool{}
			for i := 0; i < dps.Len(); i++ {
				dp := dps.At(i)
				method, _ := dp.LabelsMap().Get(metadata.Labels.Buckets)
				label := fmt.Sprintf("%s method: %s", m.Name(), method)
				metricsMap[label] = true
			}
			require.Equal(t, 2, len(metricsMap))
			require.Equal(t, map[string]bool{
				"couchbase.b_data_used method: otelb":       true,
				"couchbase.b_data_used method: test_bucket": true,
			}, metricsMap)
		case metadata.M.CouchbaseBDiskFetches.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 2, dps.Len())

			metricsMap := map[string]bool{}
			for i := 0; i < dps.Len(); i++ {
				dp := dps.At(i)
				method, _ := dp.LabelsMap().Get(metadata.Labels.Buckets)
				label := fmt.Sprintf("%s method: %s", m.Name(), method)
				metricsMap[label] = true
			}
			require.Equal(t, 2, len(metricsMap))
			require.Equal(t, map[string]bool{
				"couchbase.b_disk_fetches method: otelb":       true,
				"couchbase.b_disk_fetches method: test_bucket": true,
			}, metricsMap)
		case metadata.M.CouchbaseBDiskUsed.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 2, dps.Len())

			metricsMap := map[string]bool{}
			for i := 0; i < dps.Len(); i++ {
				dp := dps.At(i)
				method, _ := dp.LabelsMap().Get(metadata.Labels.Buckets)
				label := fmt.Sprintf("%s method: %s", m.Name(), method)
				metricsMap[label] = true
			}
			require.Equal(t, 2, len(metricsMap))
			require.Equal(t, map[string]bool{
				"couchbase.b_disk_used method: otelb":       true,
				"couchbase.b_disk_used method: test_bucket": true,
			}, metricsMap)
		case metadata.M.CouchbaseBItemCount.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 2, dps.Len())

			metricsMap := map[string]bool{}
			for i := 0; i < dps.Len(); i++ {
				dp := dps.At(i)
				method, _ := dp.LabelsMap().Get(metadata.Labels.Buckets)
				label := fmt.Sprintf("%s method: %s", m.Name(), method)
				metricsMap[label] = true
			}
			require.Equal(t, 2, len(metricsMap))
			require.Equal(t, map[string]bool{
				"couchbase.b_item_count method: otelb":       true,
				"couchbase.b_item_count method: test_bucket": true,
			}, metricsMap)
		case metadata.M.CouchbaseBMemUsed.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 2, dps.Len())

			metricsMap := map[string]bool{}
			for i := 0; i < dps.Len(); i++ {
				dp := dps.At(i)
				method, _ := dp.LabelsMap().Get(metadata.Labels.Buckets)
				label := fmt.Sprintf("%s method: %s", m.Name(), method)
				metricsMap[label] = true
			}
			require.Equal(t, 2, len(metricsMap))
			require.Equal(t, map[string]bool{
				"couchbase.b_mem_used method: otelb":       true,
				"couchbase.b_mem_used method: test_bucket": true,
			}, metricsMap)
		case metadata.M.CouchbaseBOps.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 2, dps.Len())

			metricsMap := map[string]bool{}
			for i := 0; i < dps.Len(); i++ {
				dp := dps.At(i)
				method, _ := dp.LabelsMap().Get(metadata.Labels.Buckets)
				label := fmt.Sprintf("%s method: %s", m.Name(), method)
				metricsMap[label] = true
			}
			require.Equal(t, 2, len(metricsMap))
			require.Equal(t, map[string]bool{
				"couchbase.b_ops method: otelb":       true,
				"couchbase.b_ops method: test_bucket": true,
			}, metricsMap)
		case metadata.M.CouchbaseBQuotaUsed.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 2, dps.Len())

			metricsMap := map[string]bool{}
			for i := 0; i < dps.Len(); i++ {
				dp := dps.At(i)
				method, _ := dp.LabelsMap().Get(metadata.Labels.Buckets)
				label := fmt.Sprintf("%s method: %s", m.Name(), method)
				metricsMap[label] = true
			}
			require.Equal(t, 2, len(metricsMap))
			require.Equal(t, map[string]bool{
				"couchbase.b_quota_used method: otelb":       true,
				"couchbase.b_quota_used method: test_bucket": true,
			}, metricsMap)
		case metadata.M.CouchbaseCmdGet.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
		case metadata.M.CouchbaseCPUUtilizationRate.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
		case metadata.M.CouchbaseCurrItems.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
		case metadata.M.CouchbaseCurrItemsTot.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
		case metadata.M.CouchbaseDiskFetches.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
		case metadata.M.CouchbaseGetHits.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
		case metadata.M.CouchbaseMemFree.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
		case metadata.M.CouchbaseMemTotal.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
		case metadata.M.CouchbaseMemUsed.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
		case metadata.M.CouchbaseOps.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
		case metadata.M.CouchbaseSwapTotal.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
		case metadata.M.CouchbaseSwapUsed.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
		case metadata.M.CouchbaseUptime.Name():
			dps := m.Sum().DataPoints()
			require.Equal(t, 1, dps.Len())
		default:
			require.Nil(t, m.Name(), fmt.Sprintf("metrics %s not expected", m.Name()))
		}
	}
}
