package elasticsearchreceiver

import (
	"context"
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

	expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testscraper/expected_metrics.json")
	require.NoError(t, err)

	helper.ScraperTest(t, sc.scrape, expectedFileBytes)
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
