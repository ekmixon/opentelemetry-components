package elasticsearchreceiver

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/observiq/opentelemetry-components/receiver/elasticsearchreceiver/internal/metadata"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"
)

func TestScraper(t *testing.T) {
	nodes, err := ioutil.ReadFile("./testdata/nodes.json")
	require.NoError(t, err)
	cluster, err := ioutil.ReadFile("./testdata/cluster.json")
	require.NoError(t, err)
	health, err := ioutil.ReadFile("./testdata/health.json")
	require.NoError(t, err)
	rabbitmqMock := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/_nodes/stats" {
			rw.WriteHeader(200)
			_, err = rw.Write(nodes)
			require.NoError(t, err)
			return
		}
		if req.URL.Path == "/_cluster/stats" {
			rw.WriteHeader(200)
			_, err = rw.Write(cluster)
			require.NoError(t, err)
			return
		}
		if req.URL.Path == "/_cluster/health" {
			rw.WriteHeader(200)
			_, err = rw.Write(health)
			require.NoError(t, err)
			return
		}
		rw.WriteHeader(404)
	}))
	sc, err := newElasticSearchScraper(zap.NewNop(), &Config{
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: rabbitmqMock.URL,
		},
		Username: "dev",
		Password: "dev",
	})
	require.NoError(t, err)
	err = sc.start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)
	rms, err := sc.scrape(context.Background())
	require.Nil(t, err)

	require.Equal(t, 1, rms.Len())
	rm := rms.At(0)

	ilms := rm.InstrumentationLibraryMetrics()
	require.Equal(t, 1, ilms.Len())

	ilm := ilms.At(0)
	ms := ilm.Metrics()

	require.Equal(t, 17, ms.Len())

	metricValues := make(map[string]interface{}, 7)

	for i := 0; i < ms.Len(); i++ {
		m := ms.At(i)
		name := m.Name()

		switch m.DataType() {
		case pdata.MetricDataTypeGauge:
			dps := m.Gauge().DataPoints()
			switch name {
			case "elasticsearch.cache_memory_usage":
				for i := 0; i < dps.Len(); i++ {
					dp := dps.At(i)
					dbLabel, _ := dp.LabelsMap().Get(metadata.L.CacheName)
					label := fmt.Sprintf("%s %s", m.Name(), dbLabel)
					metricValues[label] = dp.Value()
				}
			case "elasticsearch.memory_usage":
				for i := 0; i < dps.Len(); i++ {
					dp := dps.At(i)
					dbLabel, _ := dp.LabelsMap().Get(metadata.L.MemoryType)
					label := fmt.Sprintf("%s %s", m.Name(), dbLabel)
					metricValues[label] = dp.Value()
				}
			case "elasticsearch.current_documents":
				for i := 0; i < dps.Len(); i++ {
					dp := dps.At(i)
					dbLabel, _ := dp.LabelsMap().Get(metadata.L.DocumentType)
					label := fmt.Sprintf("%s %s", m.Name(), dbLabel)
					metricValues[label] = dp.Value()
				}
			case "elasticsearch.http_connections":
				for i := 0; i < dps.Len(); i++ {
					dp := dps.At(i)
					label := fmt.Sprint(m.Name())
					metricValues[label] = dp.Value()
				}
			case "elasticsearch.open_files":
				for i := 0; i < dps.Len(); i++ {
					dp := dps.At(i)
					label := fmt.Sprint(m.Name())
					metricValues[label] = dp.Value()
				}
			case "elasticsearch.server_connections":
				for i := 0; i < dps.Len(); i++ {
					dp := dps.At(i)
					label := fmt.Sprint(m.Name())
					metricValues[label] = dp.Value()
				}
			case "elasticsearch.peak_threads":
				for i := 0; i < dps.Len(); i++ {
					dp := dps.At(i)
					label := fmt.Sprint(m.Name())
					metricValues[label] = dp.Value()
				}
			case "elasticsearch.storage_size":
				for i := 0; i < dps.Len(); i++ {
					dp := dps.At(i)
					label := fmt.Sprint(m.Name())
					metricValues[label] = dp.Value()
				}
			case "elasticsearch.threads":
				for i := 0; i < dps.Len(); i++ {
					dp := dps.At(i)
					label := fmt.Sprint(m.Name())
					metricValues[label] = dp.Value()
				}
			case "elasticsearch.data_nodes":
				for i := 0; i < dps.Len(); i++ {
					dp := dps.At(i)
					label := fmt.Sprint(m.Name())
					metricValues[label] = dp.Value()
				}
			case "elasticsearch.nodes":
				for i := 0; i < dps.Len(); i++ {
					dp := dps.At(i)
					label := fmt.Sprint(m.Name())
					metricValues[label] = dp.Value()
				}
			case "elasticsearch.shards":
				for i := 0; i < dps.Len(); i++ {
					dp := dps.At(i)
					dbLabel, _ := dp.LabelsMap().Get(metadata.L.ShardType)
					label := fmt.Sprintf("%s %s", m.Name(), dbLabel)
					metricValues[label] = dp.Value()
				}
			}
		case pdata.MetricDataTypeSum:
			m.Sum().DataPoints()
			dps := m.Sum().DataPoints()
			switch name {
			case "elasticsearch.evictions":
				for i := 0; i < dps.Len(); i++ {
					dp := dps.At(i)
					dbLabel, _ := dp.LabelsMap().Get(metadata.L.CacheName)
					label := fmt.Sprintf("%s %s", m.Name(), dbLabel)
					metricValues[label] = dp.IntVal()
				}
			case "elasticsearch.gc_collection":
				for i := 0; i < dps.Len(); i++ {
					dp := dps.At(i)
					dbLabel, _ := dp.LabelsMap().Get(metadata.L.GcType)
					label := fmt.Sprintf("%s %s", m.Name(), dbLabel)
					metricValues[label] = dp.IntVal()
				}
			case "elasticsearch.network":
				for i := 0; i < dps.Len(); i++ {
					dp := dps.At(i)
					dbLabel, _ := dp.LabelsMap().Get(metadata.L.Direction)
					label := fmt.Sprintf("%s %s", m.Name(), dbLabel)
					metricValues[label] = dp.IntVal()
				}
			case "elasticsearch.operations":
				for i := 0; i < dps.Len(); i++ {
					dp := dps.At(i)
					dbLabel, _ := dp.LabelsMap().Get(metadata.L.Operation)
					label := fmt.Sprintf("%s %s", m.Name(), dbLabel)
					metricValues[label] = dp.IntVal()
				}
			case "elasticsearch.operation_time":
				for i := 0; i < dps.Len(); i++ {
					dp := dps.At(i)
					dbLabel, _ := dp.LabelsMap().Get(metadata.L.Operation)
					label := fmt.Sprintf("%s %s", m.Name(), dbLabel)
					metricValues[label] = dp.IntVal()
				}
			}
		}
	}

	require.Equal(t, map[string]interface{}{
		"elasticsearch.cache_memory_usage field":   float64(0.0),
		"elasticsearch.cache_memory_usage query":   float64(0.0),
		"elasticsearch.cache_memory_usage request": float64(0.0),
		"elasticsearch.current_documents deleted":  float64(0.0),
		"elasticsearch.current_documents live":     float64(0.0),
		"elasticsearch.data_nodes":                 float64(1.0),
		"elasticsearch.gc_collection old":          int64(0),
		"elasticsearch.gc_collection young":        int64(20),
		"elasticsearch.http_connections":           float64(2.0),
		"elasticsearch.memory_usage heap":          float64(3.05152e+08),
		"elasticsearch.memory_usage non-heap":      float64(1.28825192e+08),
		"elasticsearch.network receive":            int64(0),
		"elasticsearch.network transmit":           int64(0),
		"elasticsearch.nodes":                      float64(1.0),
		"elasticsearch.open_files":                 float64(270.0),
		"elasticsearch.operation_time delete":      int64(0),
		"elasticsearch.operation_time fetch":       int64(0),
		"elasticsearch.operation_time get":         int64(0),
		"elasticsearch.operation_time index":       int64(0),
		"elasticsearch.operation_time query":       int64(0),
		"elasticsearch.operations delete":          int64(0),
		"elasticsearch.operations fetch":           int64(0),
		"elasticsearch.operations get":             int64(0),
		"elasticsearch.operations index":           int64(0),
		"elasticsearch.operations query":           int64(0),
		"elasticsearch.peak_threads":               float64(28.0),
		"elasticsearch.server_connections":         float64(0.0),
		"elasticsearch.shards active":              float64(0.0),
		"elasticsearch.shards initializing":        float64(0.0),
		"elasticsearch.shards relocating":          float64(0.0),
		"elasticsearch.shards unassigned":          float64(0.0),
		"elasticsearch.storage_size":               float64(0.0),
		"elasticsearch.threads":                    float64(27.0),
		"elasticsearch.evictions field":            int64(0),
		"elasticsearch.evictions query":            int64(0),
		"elasticsearch.evictions request":          int64(0),
	}, metricValues)
}

func TestScraperFailedStart(t *testing.T) {
	sc, err := newElasticSearchScraper(zap.NewNop(), &Config{
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: "localhost:9200",
			TLSSetting: configtls.TLSClientSetting{
				TLSSetting: configtls.TLSSetting{
					CAFile: "/non/existent",
				},
			},
		},
		Username: "dev",
		Password: "dev",
	})
	require.NoError(t, err)
	err = sc.start(context.Background(), componenttest.NewNopHost())
	require.Error(t, err)
}
