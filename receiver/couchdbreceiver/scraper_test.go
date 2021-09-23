package couchdbreceiver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"

	"github.com/observiq/opentelemetry-components/receiver/couchdbreceiver/internal/metadata"
)

func TestScraper(t *testing.T) {
	t.Run("mocking curl response of couchbase 3.0.0", func(t *testing.T) {
		sc := newCouchdbScraper(zap.NewNop(), &Config{})
		sc.client = &fakeClient{filename: "response300.json"}

		rms, err := sc.scrape(context.Background())
		require.Nil(t, err)
		require.Equal(t, 1, rms.Len())

		rm := rms.At(0)
		ilms := rm.InstrumentationLibraryMetrics()
		require.Equal(t, 1, ilms.Len())

		ilm := ilms.At(0)
		ms := ilm.Metrics()
		require.Equal(t, len(metadata.M.Names()), ms.Len())

		validateScraperResult(t, ms)
	})
	t.Run("mocking curl response of couchbase 3.1.1", func(t *testing.T) {
		sc := newCouchdbScraper(zap.NewNop(), &Config{})
		sc.client = &fakeClient{filename: "response311.json"}

		rms, err := sc.scrape(context.Background())
		require.Nil(t, err)
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
		case metadata.M.CouchdbRequestTime.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 1.0, dps.At(0).DoubleVal())
		case metadata.M.CouchdbHttpdBulkRequests.Name():
			dps := m.Sum().DataPoints()
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 2, dps.At(0).IntVal())
		case metadata.M.CouchdbHttpdRequestMethods.Name():
			dps := m.Sum().DataPoints()
			require.Equal(t, 7, dps.Len())

			requestMethodMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				method, _ := dp.Attributes().Get(metadata.L.HTTPMethod)
				label := fmt.Sprintf("%s method:%s", m.Name(), method.AsString())
				requestMethodMetrics[label] = dp.IntVal()
			}

			require.Equal(t, 7, len(requestMethodMetrics))
			require.Equal(t, map[string]int64{
				"couchdb.httpd_request_methods method:COPY":    3,
				"couchdb.httpd_request_methods method:DELETE":  4,
				"couchdb.httpd_request_methods method:GET":     5,
				"couchdb.httpd_request_methods method:HEAD":    6,
				"couchdb.httpd_request_methods method:OPTIONS": 7,
				"couchdb.httpd_request_methods method:POST":    8,
				"couchdb.httpd_request_methods method:PUT":     9,
			},
				requestMethodMetrics)
		case metadata.M.CouchdbHttpdResponseCodes.Name():
			dps := m.Sum().DataPoints()
			require.Equal(t, 24, dps.Len())

			respondCodeMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				code, _ := dp.Attributes().Get(metadata.L.ResponseCode)
				label := fmt.Sprintf("%s code:%s", m.Name(), code.AsString())
				respondCodeMetrics[label] = dp.IntVal()
			}

			require.Equal(t, 24, len(respondCodeMetrics))
			require.Equal(t, map[string]int64{
				"couchdb.httpd_response_codes code:response_200": 10,
				"couchdb.httpd_response_codes code:response_201": 11,
				"couchdb.httpd_response_codes code:response_202": 12,
				"couchdb.httpd_response_codes code:response_204": 13,
				"couchdb.httpd_response_codes code:response_206": 14,
				"couchdb.httpd_response_codes code:response_301": 15,
				"couchdb.httpd_response_codes code:response_302": 16,
				"couchdb.httpd_response_codes code:response_304": 17,
				"couchdb.httpd_response_codes code:response_400": 18,
				"couchdb.httpd_response_codes code:response_401": 19,
				"couchdb.httpd_response_codes code:response_403": 20,
				"couchdb.httpd_response_codes code:response_404": 21,
				"couchdb.httpd_response_codes code:response_405": 22,
				"couchdb.httpd_response_codes code:response_406": 23,
				"couchdb.httpd_response_codes code:response_409": 24,
				"couchdb.httpd_response_codes code:response_412": 25,
				"couchdb.httpd_response_codes code:response_413": 26,
				"couchdb.httpd_response_codes code:response_414": 27,
				"couchdb.httpd_response_codes code:response_415": 28,
				"couchdb.httpd_response_codes code:response_416": 29,
				"couchdb.httpd_response_codes code:response_417": 30,
				"couchdb.httpd_response_codes code:response_500": 31,
				"couchdb.httpd_response_codes code:response_501": 32,
				"couchdb.httpd_response_codes code:response_503": 33,
			},
				respondCodeMetrics)
		case metadata.M.CouchdbHttpdTemporaryViewReads.Name():
			dps := m.Sum().DataPoints()
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 34, dps.At(0).IntVal())
		case metadata.M.CouchdbViewReads.Name():
			dps := m.Sum().DataPoints()
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 35, dps.At(0).IntVal())
		case metadata.M.CouchdbOpenDatabases.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 36, dps.At(0).DoubleVal())
		case metadata.M.CouchdbOpenFiles.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 37, dps.At(0).DoubleVal())
		case metadata.M.CouchdbReads.Name():
			dps := m.Sum().DataPoints()
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 38, dps.At(0).IntVal())
		case metadata.M.CouchdbWrites.Name():
			dps := m.Sum().DataPoints()
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 39, dps.At(0).IntVal())
		default:
			require.Nil(t, m.Name(), fmt.Sprintf("metrics %s not expected", m.Name()))
		}
	}
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
