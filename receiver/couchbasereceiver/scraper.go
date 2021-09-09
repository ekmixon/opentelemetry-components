package couchbasereceiver

import (
	"context"
	"errors"
	"fmt"

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

	stats, err := c.GetStats()
	if err != nil {
		c.logger.Error("Failed to fetch couchbase metrics", zap.Error(err))
		return pdata.ResourceMetricsSlice{}, errors.New("failed to fetch couchbase stats")
	}

	// metric initialization
	rms := pdata.NewResourceMetricsSlice()
	ilm := rms.AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty()
	ilm.InstrumentationLibrary().SetName("otel/couchbase")
	// now := pdata.TimestampFromTime(time.Now())

	for _, bucket := range stats.BucketsStats {
		c.logger.Info(fmt.Sprintf("%s", bucket.Name))
	}

	return rms, nil
}

func (c *couchbaseScraper) GetStats() (*Stats, error) {
	stats, err := c.client.Get()
	if err != nil {
		return nil, err
	}
	return stats, nil
}
