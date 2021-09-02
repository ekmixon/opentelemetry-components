package couchbasereceiver

import (
	"context"
	"errors"
	"strconv"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"

	"github.com/observiq/opentelemetry-components/receiver/couchbasereceiver/internal/metadata"
)

type couchbaseScraper struct {
	logger *zap.Logger
	cfg    *Config
	client client
}

func newCouchbaseScraper(logger *zap.Logger, cfg *Config) *couchbaseScraper {
	return &couchbaseScraper{
		logger: logger,
		cfg:    cfg,
	}
}

func (c *couchbaseScraper) start(ctx context.Context, host component.Host) error {
	httpClient, err := newCouchbaseClient(host, c.cfg, c.logger)
	if err != nil {
		c.logger.Error("failed to connect to couchbase", zap.Error(err))
		return err
	}
	c.client = httpClient
	return nil
}

// initMetric initializes a metric with a metadata label.
func initMetric(ms pdata.MetricSlice, mi metadata.MetricIntf) pdata.Metric {
	m := ms.AppendEmpty()
	mi.Init(m)
	return m
}

// addToDoubleMetric adds and labels a double gauge datapoint to a metricslice.
func addToDoubleMetric(metric pdata.NumberDataPointSlice, labels pdata.StringMap, value float64, ts pdata.Timestamp) {
	dataPoint := metric.AppendEmpty()
	dataPoint.SetTimestamp(ts)
	dataPoint.SetDoubleVal(value)
	if labels.Len() > 0 {
		labels.CopyTo(dataPoint.LabelsMap())
	}
}

// addToIntMetric adds and labels a int sum datapoint to metricslice.
func addToIntMetric(metric pdata.NumberDataPointSlice, labels pdata.StringMap, value int64, ts pdata.Timestamp) {
	dataPoint := metric.AppendEmpty()
	dataPoint.SetTimestamp(ts)
	dataPoint.SetIntVal(value)
	if labels.Len() > 0 {
		labels.CopyTo(dataPoint.LabelsMap())
	}
}

func (c *couchbaseScraper) scrape(context.Context) (pdata.ResourceMetricsSlice, error) {
	if c.client == nil {
		return pdata.ResourceMetricsSlice{}, errors.New("failed to connect to couchbase client")
	}

	metricsResults, err := c.GetMetrics()
	if err != nil {
		c.logger.Error("Failed to fetch couchbase metrics", zap.Error(err))
		return pdata.ResourceMetricsSlice{}, errors.New("failed to fetch couchbase stats")
	}

	// metric initialization
	rms := pdata.NewResourceMetricsSlice()
	ilm := rms.AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty()
	ilm.InstrumentationLibrary().SetName("otel/couchbase")
	now := pdata.TimestampFromTime(time.Now())

	// c.logger.Info(fmt.Sprintf("METRICS GOT %v", metricsResults))
	for i := 0; i < len(metricsResults); i++ {
		metricResult := metricsResults[i]
		// c.logger.Info(fmt.Sprintf("Processing Metric %v", metricResult))
		if metricResult.Kind == "gauge" {
			gaugeSlice := initMetric(ilm.Metrics(), metricResult.MetadataLabel).Gauge().DataPoints()
			labels := pdata.NewStringMap()
			for _, metricLabel := range metricResult.Labels {
				labels.Insert(metricResult.MetadataLabel.Name(), metricLabel)
			}
			if metricResult.ValueType == "float" {
				value, ok := parseFloat(metricResult.Value)
				if !ok {
					c.logger.Info(
						err.Error(),
						zap.String("metric", metricResult.Name),
					)
				} else {
					addToDoubleMetric(gaugeSlice, labels, value, now)
				}
			} else if metricResult.ValueType == "int" {
				value, ok := parseInt(metricResult.Value)
				if !ok {
					c.logger.Info(
						err.Error(),
						zap.String("metric", metricResult.Name),
					)
				} else {
					addToIntMetric(gaugeSlice, labels, value, now)
				}
			} else {
				c.logger.Info(
					err.Error(),
					zap.String("metric", metricResult.Name),
				)
			}

		} else if metricResult.Kind == "sum" {
			sumSlice := initMetric(ilm.Metrics(), metricResult.MetadataLabel).Sum().DataPoints()
			labels := pdata.NewStringMap()
			for _, metricLabel := range metricResult.Labels {
				labels.Insert(metricResult.MetadataLabel.Name(), metricLabel)
			}
			if metricResult.ValueType == "float" {
				value, ok := parseFloat(metricResult.Value)
				if !ok {
					c.logger.Info(
						err.Error(),
						zap.String("metric", metricResult.Name),
					)
				} else {
					addToDoubleMetric(sumSlice, labels, value, now)
				}
			} else if metricResult.ValueType == "int" {
				value, ok := parseInt(metricResult.Value)
				if !ok {
					c.logger.Info(
						err.Error(),
						zap.String("metric", metricResult.Name),
					)
				} else {
					addToIntMetric(sumSlice, labels, value, now)
				}
			} else {
				c.logger.Info(
					err.Error(),
					zap.String("metric", metricResult.Name),
				)
			}
		} else {
			continue
		}
	}

	return rms, nil
}

func (c *couchbaseScraper) GetMetrics() ([]Metric, error) {
	metrics := Metrics

	err := c.client.Post(metrics)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

// parseFloat converts string to float64.
func parseFloat(value interface{}) (float64, bool) {
	switch f := value.(type) {
	case float64:
		return f, true
	case int64:
		return float64(f), true
	case float32:
		return float64(f), true
	case int32:
		return float64(f), true
	case string:
		fConv, err := strconv.ParseFloat(f, 64)
		if err != nil {
			return 0, false
		}
		return fConv, true
	}
	return 0, false
}

func parseInt(value interface{}) (int64, bool) {
	switch i := value.(type) {
	case int64:
		return i, true
	case float64:
		return int64(i), true
	case float32:
		return int64(i), true
	case int32:
		return int64(i), true
	case string:
		intConv, err := strconv.ParseInt(i, 10, 64)
		if err != nil {
			return 0, false
		}
		return intConv, true
	}
	return 0, false
}
