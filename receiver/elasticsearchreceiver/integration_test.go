//go:build integration
// +build integration

package elasticsearchreceiver

import (
	"context"
	"fmt"
	"net"
	"path"
	"testing"
	"time"

	"github.com/observiq/opentelemetry-components/receiver/elasticsearchreceiver/internal/metadata"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/model/pdata"
)

func TestElasticSearch7_8(t *testing.T) {
	container := getContainer(t, containerRequest7_8)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = fmt.Sprintf("http://%s", net.JoinHostPort(hostname, "9200"))

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

func TestElasticSearch7_13(t *testing.T) {
	container := getContainer(t, containerRequest7_13)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = fmt.Sprintf("http://%s", net.JoinHostPort(hostname, "9200"))

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
	containerRequest7_8 = testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.elasticsearch_7.8",
		},
		ExposedPorts: []string{"9200:9200"},
		WaitingFor: wait.ForListeningPort("9200").
			WithStartupTimeout(2 * time.Minute),
	}
	containerRequest7_13 = testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.elasticsearch_7.13",
		},
		ExposedPorts: []string{"9200:9200"},
		WaitingFor: wait.ForListeningPort("9200").
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
	time.Sleep(time.Second * 6) // TODO custom wait.Strategy
	return container
}

func validateResult(t *testing.T, metrics pdata.MetricSlice) {
	require.Equal(t, len(metadata.M.Names()), metrics.Len())
	exists := make(map[string]bool)

	unenumAttributeSet := []string{
		metadata.L.ServerName,
	}

	enumAttributeSet := []string{
		metadata.L.CacheName,
		metadata.L.GcType,
		metadata.L.Direction,
		metadata.L.DocumentType,
		metadata.L.ShardType,
		metadata.L.MemoryType,
		metadata.L.Operation,
		metadata.L.ThreadType,
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

			for _, attribute := range unenumAttributeSet {
				_, ok := dp.Attributes().Get(attribute)
				if ok {
					key = key + " " + attribute
				}
			}

			for _, attribute := range enumAttributeSet {
				attributeVal, ok := dp.Attributes().Get(attribute)
				if ok {
					key += " " + attributeVal.AsString()
				}
			}
			exists[key] = true
		}
	}
	require.Equal(t, map[string]bool{
		"elasticsearch.cache_memory_usage server_name field":   true,
		"elasticsearch.cache_memory_usage server_name query":   true,
		"elasticsearch.cache_memory_usage server_name request": true,
		"elasticsearch.evictions server_name field":            true,
		"elasticsearch.evictions server_name query":            true,
		"elasticsearch.evictions server_name request":          true,
		"elasticsearch.gc_collection server_name young":        true,
		"elasticsearch.gc_collection_time server_name young":   true,
		"elasticsearch.gc_collection server_name old":          true,
		"elasticsearch.gc_collection_time server_name old":     true,
		"elasticsearch.memory_usage server_name heap":          true,
		"elasticsearch.memory_usage server_name non-heap":      true,
		"elasticsearch.network server_name transmit":           true,
		"elasticsearch.network server_name receive":            true,
		"elasticsearch.current_documents server_name live":     true,
		"elasticsearch.current_documents server_name deleted":  true,
		"elasticsearch.http_connections server_name":           true,
		"elasticsearch.open_files server_name":                 true,
		"elasticsearch.server_connections server_name":         true,
		"elasticsearch.operations server_name get":             true,
		"elasticsearch.operations server_name delete":          true,
		"elasticsearch.operations server_name index":           true,
		"elasticsearch.operations server_name query":           true,
		"elasticsearch.operations server_name fetch":           true,
		"elasticsearch.operation_time server_name get":         true,
		"elasticsearch.operation_time server_name delete":      true,
		"elasticsearch.operation_time server_name index":       true,
		"elasticsearch.operation_time server_name query":       true,
		"elasticsearch.operation_time server_name fetch":       true,
		"elasticsearch.peak_threads server_name":               true,
		"elasticsearch.storage_size server_name":               true,
		"elasticsearch.threads server_name":                    true,
		"elasticsearch.thread_pools server_name active":        true,
		"elasticsearch.thread_pools server_name completed":     true,
		"elasticsearch.thread_pools server_name largest":       true,
		"elasticsearch.thread_pools server_name queue":         true,
		"elasticsearch.thread_pools server_name rejected":      true,
		"elasticsearch.thread_pools server_name total":         true,
		"elasticsearch.data_nodes":                             true,
		"elasticsearch.nodes":                                  true,
		"elasticsearch.shards unassigned":                      true,
		"elasticsearch.shards active":                          true,
		"elasticsearch.shards relocating":                      true,
		"elasticsearch.shards initializing":                    true,
	}, exists)
}
