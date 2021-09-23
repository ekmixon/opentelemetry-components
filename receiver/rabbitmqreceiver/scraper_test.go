package rabbitmqreceiver

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/observiq/opentelemetry-components/receiver/rabbitmqreceiver/internal/metadata"
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
	rms, err := sc.scrape(context.Background())
	require.Nil(t, err)

	require.Equal(t, 1, rms.Len())
	rm := rms.At(0)

	ilms := rm.InstrumentationLibraryMetrics()
	require.Equal(t, 1, ilms.Len())

	ilm := ilms.At(0)
	ms := ilm.Metrics()

	require.Equal(t, 4, ms.Len())

	metricValues := make(map[string]float64, 7)

	for i := 0; i < ms.Len(); i++ {
		m := ms.At(i)

		dps := m.Gauge().DataPoints()
		if dps.Len() > 1 {
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.Attributes().Get(metadata.L.State)
				label := fmt.Sprintf("%s state:%s", m.Name(), state.AsString())
				metricValues[label] = dp.DoubleVal()
			}
		} else {
			dp := dps.At(0)
			metricValues[m.Name()] = dp.DoubleVal()
		}
	}

	require.Equal(t, map[string]float64{
		"rabbitmq.publish_rate":                      1,
		"rabbitmq.delivery_rate":                     1.4,
		"rabbitmq.consumers":                         1,
		"rabbitmq.num_messages state:ready":          6,
		"rabbitmq.num_messages state:total":          7,
		"rabbitmq.num_messages state:unacknowledged": 1,
	}, metricValues)
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
