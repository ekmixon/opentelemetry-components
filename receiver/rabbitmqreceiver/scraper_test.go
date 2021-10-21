package rabbitmqreceiver

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
	f, err := ioutil.ReadFile("./testdata/exampleAPICall.json")
	require.NoError(t, err)
	rabbitmqMock := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/api/queues" {
			rw.WriteHeader(200)
			_, err = rw.Write(f)
			require.NoError(t, err)
			return
		}
		rw.WriteHeader(404)
	}))
	sc, err := newRabbitMQScraper(zap.NewNop(), &Config{
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
	sc, err := newRabbitMQScraper(zap.NewNop(), &Config{
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: "localhost:8080",
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
