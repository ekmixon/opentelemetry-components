package couchdbreceiver

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
	t.Run("mocking curl response of couchbase 2.3.1", func(t *testing.T) {
		sc := newCouchdbScraper(zap.NewNop(), &Config{})
		sc.client = &fakeClient{filename: "response231.json"}

		expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testscraper/expected_metrics.json")
		require.NoError(t, err)

		helper.ScraperTest(t, sc.scrape, expectedFileBytes)
	})
	t.Run("mocking curl response of couchbase 3.1.2", func(t *testing.T) {
		sc := newCouchdbScraper(zap.NewNop(), &Config{})
		sc.client = &fakeClient{filename: "response312.json"}

		expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testscraper/expected_metrics.json")
		require.NoError(t, err)

		helper.ScraperTest(t, sc.scrape, expectedFileBytes)
	})
}

func TestScraperError(t *testing.T) {
	t.Run("no client", func(t *testing.T) {
		sc := newCouchdbScraper(zap.NewNop(), &Config{})
		sc.client = nil

		_, err := sc.scrape(context.Background())
		require.Error(t, err)
		require.EqualValues(t, errors.New("failed to connect to couchdb client"), err)
	})
	t.Run("error on get", func(t *testing.T) {
		sc := newCouchdbScraper(zap.NewNop(), &Config{})
		sc.client = &fakeClient{err: errors.New("failed to fetch couchdb stats")}

		_, err := sc.scrape(context.Background())
		require.Error(t, err)
		require.EqualValues(t, errors.New("failed to fetch couchdb stats"), err)
	})
}

func TestStart(t *testing.T) {
	t.Run("failed scrape", func(t *testing.T) {
		sc := newCouchdbScraper(zap.NewNop(), &Config{
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
		couchdbMock := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.URL.Path == "/_node/_local/_stats/couchdb" {
				rw.WriteHeader(200)
				_, _ = rw.Write([]byte(``))
				return
			}
			rw.WriteHeader(404)
		}))
		sc := newCouchdbScraper(zap.NewNop(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: couchdbMock.URL + "/_node/_local/_stats/couchdb",
			},
		})
		err := sc.start(context.Background(), componenttest.NewNopHost())
		require.NoError(t, err)
	})
}
