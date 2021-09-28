package couchdbreceiver

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/model/otlp"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"
)

func TestScraper(t *testing.T) {
	t.Run("mocking curl response of couchbase 3.0.0", func(t *testing.T) {
		sc := newCouchdbScraper(zap.NewNop(), &Config{})
		sc.client = &fakeClient{filename: "response300.json"}

		actualMetrics := pdata.NewMetrics()
		rms := actualMetrics.ResourceMetrics()
		scrapedRMS, err := sc.scrape(context.Background())
		require.NoError(t, err)
		scrapedRMS.CopyTo(rms)

		expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testscraper/expected_metrics.json")
		require.NoError(t, err)
		unmarshaller := otlp.NewJSONMetricsUnmarshaler()
		expectedMetrics, err := unmarshaller.UnmarshalMetrics(expectedFileBytes)
		require.NoError(t, err)

		aMetricSlice := expectedMetrics.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics()
		eMetricSlice := actualMetrics.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics()

		require.NoError(t, compareMetrics(eMetricSlice, aMetricSlice))
	})
	t.Run("mocking curl response of couchbase 3.1.1", func(t *testing.T) {
		sc := newCouchdbScraper(zap.NewNop(), &Config{})
		sc.client = &fakeClient{filename: "response311.json"}

		actualMetrics := pdata.NewMetrics()
		rms := actualMetrics.ResourceMetrics()
		scrapedRMS, err := sc.scrape(context.Background())
		require.NoError(t, err)
		scrapedRMS.CopyTo(rms)

		expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testscraper/expected_metrics.json")
		require.NoError(t, err)
		unmarshaller := otlp.NewJSONMetricsUnmarshaler()
		expectedMetrics, err := unmarshaller.UnmarshalMetrics(expectedFileBytes)
		require.NoError(t, err)

		aMetricSlice := expectedMetrics.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics()
		eMetricSlice := actualMetrics.ResourceMetrics().At(0).InstrumentationLibraryMetrics().At(0).Metrics()

		require.NoError(t, compareMetrics(eMetricSlice, aMetricSlice))
	})
}

func compareMetrics(expectedAll, actualAll pdata.MetricSlice) error {
	if actualAll.Len() != expectedAll.Len() {
		return fmt.Errorf("metrics not of same length")
	}

	lessFunc := func(a, b pdata.Metric) bool {
		return a.Name() < b.Name()
	}

	actualMetrics := actualAll.Sort(lessFunc)
	expectedMetrics := expectedAll.Sort(lessFunc)

	for i := 0; i < actualMetrics.Len(); i++ {
		actual := actualMetrics.At(i)
		expected := expectedMetrics.At(i)

		if actual.Name() != expected.Name() {
			return fmt.Errorf("metric name does not match expected: %s, actual: %s", expected.Name(), actual.Name())
		}
		if actual.DataType() != expected.DataType() {
			return fmt.Errorf("metric datatype does not match expected: %s, actual: %s", expected.DataType(), actual.DataType())
		}
		if actual.Description() != expected.Description() {
			return fmt.Errorf("metric description does not match expected: %s, actual: %s", expected.Description(), actual.Description())
		}
		if actual.Unit() != expected.Unit() {
			return fmt.Errorf("metric Unit does not match expected: %s, actual: %s", expected.Unit(), actual.Unit())
		}

		var actualDataPoints pdata.NumberDataPointSlice
		var expectedDataPoints pdata.NumberDataPointSlice

		switch actual.DataType() {
		case pdata.MetricDataTypeGauge:
			actualDataPoints = actual.Gauge().DataPoints()
			expectedDataPoints = expected.Gauge().DataPoints()
		case pdata.MetricDataTypeSum:
			if actual.Sum().AggregationTemporality() != expected.Sum().AggregationTemporality() {
				return fmt.Errorf("metric AggregationTemporality does not match expected: %s, actual: %s", expected.Sum().AggregationTemporality(), actual.Sum().AggregationTemporality())
			}
			if actual.Sum().IsMonotonic() != expected.Sum().IsMonotonic() {
				return fmt.Errorf("metric IsMonotonic does not match expected: %t, actual: %t", expected.Sum().IsMonotonic(), actual.Sum().IsMonotonic())
			}
			actualDataPoints = actual.Sum().DataPoints()
			expectedDataPoints = expected.Sum().DataPoints()
		}

		if actualDataPoints.Len() != expectedDataPoints.Len() {
			return fmt.Errorf("length of datapoints don't match")
		}

		dataPointMatches := 0
		for j := 0; j < expectedDataPoints.Len(); j++ {
			edp := expectedDataPoints.At(j)
			for k := 0; k < actualDataPoints.Len(); k++ {
				adp := actualDataPoints.At(k)
				adpAttributes := adp.Attributes()
				labelMatches := true

				if edp.Attributes().Len() != adpAttributes.Len() {
					break
				}
				edp.Attributes().Range(func(k string, v pdata.AttributeValue) bool {
					if attributeVal, ok := adpAttributes.Get(k); ok && attributeVal.StringVal() == v.StringVal() {
						return true
					}
					labelMatches = false
					return false
				})
				if !labelMatches {
					continue
				}
				if edp.IntVal() != adp.IntVal() {
					return fmt.Errorf("metric datapoint IntVal doesn't match expected: %d, actual: %d", edp.IntVal(), adp.IntVal())
				}
				if edp.DoubleVal() != adp.DoubleVal() {
					return fmt.Errorf("metric datapoint DoubleVal doesn't match expected: %f, actual: %f", edp.DoubleVal(), adp.DoubleVal())
				}
				dataPointMatches++
				break
			}
		}
		if dataPointMatches != expectedDataPoints.Len() {
			return fmt.Errorf("missing Datapoints")
		}
	}
	return nil
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
