package couchbasereceiver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/observiq/opentelemetry-components/receiver/couchbasereceiver/internal/metadata"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"
)

func TestScraper(t *testing.T) {
	t.Run("mocking curl response of couchbase 6.6", func(t *testing.T) {
		sc := newCouchbaseScraper(zap.NewNop(), &Config{})
		sc.client = &fakeClient{
			nodeStatsFilename:    "node_stats_6_6.json",
			bucketsStatsFilename: "buckets_stats_6_6.json"}

		stats, err := sc.client.Get()
		require.NoError(t, err)
		require.NotNil(t, stats)

		rms, err := sc.scrape(context.Background())
		require.NoError(t, err)
		require.Equal(t, 1, rms.Len())

		rm := rms.At(0)
		ilms := rm.InstrumentationLibraryMetrics()
		require.Equal(t, 1, ilms.Len())

		ilm := ilms.At(0)
		ms := ilm.Metrics()
		require.Equal(t, len(metadata.M.Names()), ms.Len())

		validateScraperResult(t, ms)
	})
	t.Run("mocking curl response of couchbase 7.0", func(t *testing.T) {
		sc := newCouchbaseScraper(zap.NewNop(), &Config{})
		sc.client = &fakeClient{
			nodeStatsFilename:    "node_stats_7_0.json",
			bucketsStatsFilename: "buckets_stats_7_0.json"}

		stats, err := sc.client.Get()
		require.NoError(t, err)
		require.NotNil(t, stats)

		rms, err := sc.scrape(context.Background())
		require.NoError(t, err)
		require.Equal(t, 1, rms.Len())

		rm := rms.At(0)
		ilms := rm.InstrumentationLibraryMetrics()
		require.Equal(t, 1, ilms.Len())

		ilm := ilms.At(0)
		ms := ilm.Metrics()
		require.Equal(t, len(metadata.M.Names()), ms.Len())

		validateScraperResult(t, ms)
	})
}

func validateScraperResult(t *testing.T, metric pdata.MetricSlice) {
	require.Equal(t, len(metadata.M.Names()), metric.Len())
	for i := 0; i < metric.Len(); i++ {
		m := metric.At(i)
		switch m.Name() {
		case metadata.M.CouchbaseBDataUsed.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 2, dps.Len())

			metricsMap := map[string]int64{}
			for i := 0; i < dps.Len(); i++ {
				dp := dps.At(i)
				method, _ := dp.LabelsMap().Get(metadata.Labels.Buckets)
				label := fmt.Sprintf("%s method: %s", m.Name(), method)
				metricsMap[label] = dp.IntVal()
			}
			require.Equal(t, 2, len(metricsMap))
			require.Equal(t, map[string]int64{
				"couchbase.b_data_used method: otelb":       4,
				"couchbase.b_data_used method: test_bucket": 8,
			}, metricsMap)
		case metadata.M.CouchbaseBDiskFetches.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 2, dps.Len())

			metricsMap := map[string]float64{}
			for i := 0; i < dps.Len(); i++ {
				dp := dps.At(i)
				method, _ := dp.LabelsMap().Get(metadata.Labels.Buckets)
				label := fmt.Sprintf("%s method: %s", m.Name(), method)
				metricsMap[label] = dp.DoubleVal()
			}
			require.Equal(t, 2, len(metricsMap))
			require.Equal(t, map[string]float64{
				"couchbase.b_disk_fetches method: otelb":       1.3,
				"couchbase.b_disk_fetches method: test_bucket": 2.3,
			}, metricsMap)
		case metadata.M.CouchbaseBDiskUsed.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 2, dps.Len())

			metricsMap := map[string]int64{}
			for i := 0; i < dps.Len(); i++ {
				dp := dps.At(i)
				method, _ := dp.LabelsMap().Get(metadata.Labels.Buckets)
				label := fmt.Sprintf("%s method: %s", m.Name(), method)
				metricsMap[label] = dp.IntVal()
			}
			require.Equal(t, 2, len(metricsMap))
			require.Equal(t, map[string]int64{
				"couchbase.b_disk_used method: otelb":       3,
				"couchbase.b_disk_used method: test_bucket": 7,
			}, metricsMap)
		case metadata.M.CouchbaseBItemCount.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 2, dps.Len())

			metricsMap := map[string]int64{}
			for i := 0; i < dps.Len(); i++ {
				dp := dps.At(i)
				method, _ := dp.LabelsMap().Get(metadata.Labels.Buckets)
				label := fmt.Sprintf("%s method: %s", m.Name(), method)
				metricsMap[label] = dp.IntVal()
			}
			require.Equal(t, 2, len(metricsMap))
			require.Equal(t, map[string]int64{
				"couchbase.b_item_count method: otelb":       2,
				"couchbase.b_item_count method: test_bucket": 6,
			}, metricsMap)
		case metadata.M.CouchbaseBMemUsed.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 2, dps.Len())

			metricsMap := map[string]int64{}
			for i := 0; i < dps.Len(); i++ {
				dp := dps.At(i)
				method, _ := dp.LabelsMap().Get(metadata.Labels.Buckets)
				label := fmt.Sprintf("%s method: %s", m.Name(), method)
				metricsMap[label] = dp.IntVal()
			}
			require.Equal(t, 2, len(metricsMap))
			require.Equal(t, map[string]int64{
				"couchbase.b_mem_used method: otelb":       5,
				"couchbase.b_mem_used method: test_bucket": 9,
			}, metricsMap)
		case metadata.M.CouchbaseBOps.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 2, dps.Len())

			metricsMap := map[string]float64{}
			for i := 0; i < dps.Len(); i++ {
				dp := dps.At(i)
				method, _ := dp.LabelsMap().Get(metadata.Labels.Buckets)
				label := fmt.Sprintf("%s method: %s", m.Name(), method)
				metricsMap[label] = dp.DoubleVal()
			}
			require.Equal(t, 2, len(metricsMap))
			require.Equal(t, map[string]float64{
				"couchbase.b_ops method: otelb":       1.2,
				"couchbase.b_ops method: test_bucket": 2.2,
			}, metricsMap)
		case metadata.M.CouchbaseBQuotaUsed.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 2, dps.Len())

			metricsMap := map[string]float64{}
			for i := 0; i < dps.Len(); i++ {
				dp := dps.At(i)
				method, _ := dp.LabelsMap().Get(metadata.Labels.Buckets)
				label := fmt.Sprintf("%s method: %s", m.Name(), method)
				metricsMap[label] = dp.DoubleVal()
			}
			require.Equal(t, 2, len(metricsMap))
			require.Equal(t, map[string]float64{
				"couchbase.b_quota_used method: otelb":       1.1,
				"couchbase.b_quota_used method: test_bucket": 2.1,
			}, metricsMap)
		case metadata.M.CouchbaseCmdGet.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 1.3, dps.At(0).DoubleVal())
		case metadata.M.CouchbaseCPUUtilizationRate.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 1.1, dps.At(0).DoubleVal())
		case metadata.M.CouchbaseCurrItems.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 5, dps.At(0).IntVal())
		case metadata.M.CouchbaseCurrItemsTot.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 6, dps.At(0).IntVal())
		case metadata.M.CouchbaseDiskFetches.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 1.2, dps.At(0).DoubleVal())
		case metadata.M.CouchbaseGetHits.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 1.4, dps.At(0).DoubleVal())
		case metadata.M.CouchbaseMemFree.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 4, dps.At(0).IntVal())
		case metadata.M.CouchbaseMemTotal.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 3, dps.At(0).IntVal())
		case metadata.M.CouchbaseMemUsed.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 7, dps.At(0).IntVal())
		case metadata.M.CouchbaseOps.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 1.5, dps.At(0).DoubleVal())
		case metadata.M.CouchbaseSwapTotal.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 1, dps.At(0).IntVal())
		case metadata.M.CouchbaseSwapUsed.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 2, dps.At(0).IntVal())
		case metadata.M.CouchbaseUptime.Name():
			dps := m.Sum().DataPoints()
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 10, dps.At(0).IntVal())
		default:
			require.Nil(t, m.Name(), fmt.Sprintf("metrics %s not expected", m.Name()))
		}
	}
}

func TestScraperError(t *testing.T) {
	t.Run("no client", func(t *testing.T) {
		sc := newCouchbaseScraper(zap.NewNop(), &Config{})
		sc.client = nil

		_, err := sc.scrape(context.Background())
		require.Error(t, err)
		require.EqualValues(t, errors.New("failed to connect to couchbase client"), err)
	})
	t.Run("error on get", func(t *testing.T) {
		sc := newCouchbaseScraper(zap.NewNop(), &Config{})
		sc.client = &fakeClient{err: errors.New("failed to fetch couchbase stats")}

		_, err := sc.scrape(context.Background())
		require.Error(t, err)
		require.EqualValues(t, errors.New("failed to fetch couchbase stats"), err)
	})
	t.Run("empty stats", func(t *testing.T) {
		sc := newCouchbaseScraper(zap.NewNop(), &Config{})
		sc.client = &fakeClient{nodeStatsFilename: "empty"}

		_, err := sc.scrape(context.Background())
		require.Error(t, err)
		require.EqualValues(t, errors.New("failed to fetch couchbase stats"), err)
	})
}

func TestStart(t *testing.T) {
	t.Run("failed scrape", func(t *testing.T) {
		sc := newCouchbaseScraper(zap.NewNop(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: "",
				TLSSetting: configtls.TLSClientSetting{
					TLSSetting: configtls.TLSSetting{
						CAFile: "/non/existent",
					},
				},
			},
		})
		err := sc.start(context.Background(), componenttest.NewNopHost())
		require.Error(t, err)
	})

	t.Run("no error", func(t *testing.T) {
		couchbaseMock := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.URL.Path == "/pools/default" {
				rw.WriteHeader(200)
				_, _ = rw.Write([]byte(``))
				return
			}
			rw.WriteHeader(404)
		}))
		sc := newCouchbaseScraper(zap.NewNop(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: couchbaseMock.URL + "/pools/default",
			},
		})
		err := sc.start(context.Background(), componenttest.NewNopHost())
		require.NoError(t, err)
	})
}

func TestMissingMetrics(t *testing.T) {
	sc := newCouchbaseScraper(zap.NewNop(), &Config{})
	sc.client = &fakeClient{
		nodeStatsFilename:    "missing_node_stats.json",
		bucketsStatsFilename: "missing_buckets_stats.json"}

	stats, err := sc.client.Get()
	require.NoError(t, err)
	require.NotNil(t, stats)

	rms, err := sc.scrape(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, rms.Len())
}
