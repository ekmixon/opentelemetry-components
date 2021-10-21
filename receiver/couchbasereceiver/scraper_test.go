package couchbasereceiver

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/observiq/opentelemetry-components/receiver/helper"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configtls"
	"go.uber.org/zap"
)

func TestScraper(t *testing.T) {
	t.Run("mocking curl response of couchbase 6.6", func(t *testing.T) {
		sc := newCouchbaseScraper(zap.NewNop(), &Config{})
		sc.client = &fakeClient{
			nodeStatsFilename:    "node_stats_6_6.json",
			bucketsStatsFilename: "buckets_stats_6_6.json"}

		expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testscraper/expected_metrics.json")
		require.NoError(t, err)

		helper.ScraperTest(t, sc.scrape, expectedFileBytes)
	})
	t.Run("mocking curl response of couchbase 7.0", func(t *testing.T) {
		sc := newCouchbaseScraper(zap.NewNop(), &Config{})
		sc.client = &fakeClient{
			nodeStatsFilename:    "node_stats_7_0.json",
			bucketsStatsFilename: "buckets_stats_7_0.json"}

		expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testscraper/expected_metrics.json")
		require.NoError(t, err)

		helper.ScraperTest(t, sc.scrape, expectedFileBytes)
	})
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
