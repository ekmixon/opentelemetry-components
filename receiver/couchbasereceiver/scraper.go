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

	stats, err := c.GetStats()
	if err != nil {
		c.logger.Error("Failed to fetch couchbase metrics", zap.Error(err))
		return pdata.ResourceMetricsSlice{}, errors.New("failed to fetch couchbase stats")
	}

	// metric initialization
	rms := pdata.NewResourceMetricsSlice()
	ilm := rms.AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty()
	ilm.InstrumentationLibrary().SetName("otel/couchbase")
	now := pdata.TimestampFromTime(time.Now())

	bDataUsed := initMetric(ilm.Metrics(), metadata.M.CouchbaseBDataUsed).Gauge().DataPoints()
	bDiskFetches := initMetric(ilm.Metrics(), metadata.M.CouchbaseBDiskFetches).Gauge().DataPoints()
	bDiskUsed := initMetric(ilm.Metrics(), metadata.M.CouchbaseBDiskUsed).Gauge().DataPoints()
	bItemCount := initMetric(ilm.Metrics(), metadata.M.CouchbaseBItemCount).Gauge().DataPoints()
	bMemUsed := initMetric(ilm.Metrics(), metadata.M.CouchbaseBMemUsed).Gauge().DataPoints()
	bOps := initMetric(ilm.Metrics(), metadata.M.CouchbaseBOps).Gauge().DataPoints()
	bQuotaUsed := initMetric(ilm.Metrics(), metadata.M.CouchbaseBQuotaUsed).Gauge().DataPoints()
	cmdGet := initMetric(ilm.Metrics(), metadata.M.CouchbaseCmdGet).Gauge().DataPoints()
	cpuUtilizationRate := initMetric(ilm.Metrics(), metadata.M.CouchbaseCPUUtilizationRate).Gauge().DataPoints()
	currItems := initMetric(ilm.Metrics(), metadata.M.CouchbaseCurrItems).Gauge().DataPoints()
	currItemsTot := initMetric(ilm.Metrics(), metadata.M.CouchbaseCurrItemsTot).Gauge().DataPoints()
	diskFetches := initMetric(ilm.Metrics(), metadata.M.CouchbaseDiskFetches).Gauge().DataPoints()
	getHits := initMetric(ilm.Metrics(), metadata.M.CouchbaseGetHits).Gauge().DataPoints()
	memFree := initMetric(ilm.Metrics(), metadata.M.CouchbaseMemFree).Gauge().DataPoints()
	memTotal := initMetric(ilm.Metrics(), metadata.M.CouchbaseMemTotal).Gauge().DataPoints()
	memUsed := initMetric(ilm.Metrics(), metadata.M.CouchbaseMemUsed).Gauge().DataPoints()
	ops := initMetric(ilm.Metrics(), metadata.M.CouchbaseOps).Gauge().DataPoints()
	swapTotal := initMetric(ilm.Metrics(), metadata.M.CouchbaseSwapTotal).Gauge().DataPoints()
	swapUsed := initMetric(ilm.Metrics(), metadata.M.CouchbaseSwapUsed).Gauge().DataPoints()
	uptime := initMetric(ilm.Metrics(), metadata.M.CouchbaseUptime).Sum().DataPoints()

	// bDataUsed
	for _, bucket := range stats.BucketsStats {
		bDataUsedLabels := pdata.NewStringMap()
		bDataUsedLabels.Insert(metadata.Labels.Buckets, bucket.Name)
		bDataUsedValues := bucket.BasicStats.DataUsed
		if bDataUsedValues == nil {
			c.logger.Info(
				"failed to collect metric",
				zap.String("metric", "bDataUsed"),
			)
		} else {
			addToIntMetric(bDataUsed, bDataUsedLabels, *bDataUsedValues, now)
		}
	}

	// bDiskFetches
	for _, bucket := range stats.BucketsStats {
		bDiskFetchesLabels := pdata.NewStringMap()
		bDiskFetchesLabels.Insert(metadata.Labels.Buckets, bucket.Name)
		bDiskFetchesValues := bucket.BasicStats.DiskFetches
		if bDiskFetchesValues == nil {
			c.logger.Info(
				"failed to collect metric",
				zap.String("metric", "bDiskFetches"),
			)
		} else {
			addToDoubleMetric(bDiskFetches, bDiskFetchesLabels, *bDiskFetchesValues, now)
		}
	}

	// bDiskUsed
	for _, bucket := range stats.BucketsStats {
		bDiskUsedLabels := pdata.NewStringMap()
		bDiskUsedLabels.Insert(metadata.Labels.Buckets, bucket.Name)
		bDiskUsedValues := bucket.BasicStats.DiskUsed
		if bDiskUsedValues == nil {
			c.logger.Info(
				"failed to collect metric",
				zap.String("metric", "bDiskUsed"),
			)
		} else {
			addToIntMetric(bDiskUsed, bDiskUsedLabels, *bDiskUsedValues, now)
		}
	}

	// bItemCount
	for _, bucket := range stats.BucketsStats {
		bItemCountLabels := pdata.NewStringMap()
		bItemCountLabels.Insert(metadata.Labels.Buckets, bucket.Name)
		bItemCountValues := bucket.BasicStats.ItemCount
		if bItemCountValues == nil {
			c.logger.Info(
				"failed to collect metric",
				zap.String("metric", "bItemCount"),
			)
		} else {
			addToIntMetric(bItemCount, bItemCountLabels, *bItemCountValues, now)
		}
	}

	// bMemUsed
	for _, bucket := range stats.BucketsStats {
		bMemUsedLabels := pdata.NewStringMap()
		bMemUsedLabels.Insert(metadata.Labels.Buckets, bucket.Name)
		bMemUsedValues := bucket.BasicStats.MemUsed
		if bMemUsedValues == nil {
			c.logger.Info(
				"failed to collect metric",
				zap.String("metric", "bMemUsed"),
			)
		} else {
			addToIntMetric(bMemUsed, bMemUsedLabels, *bMemUsedValues, now)
		}
	}
	// bOps
	for _, bucket := range stats.BucketsStats {
		bOpsLabels := pdata.NewStringMap()
		bOpsLabels.Insert(metadata.Labels.Buckets, bucket.Name)
		bOpsValues := bucket.BasicStats.OpsPerSec
		if bOpsValues == nil {
			c.logger.Info(
				"failed to collect metric",
				zap.String("metric", "bOps"),
			)
		} else {
			addToDoubleMetric(bOps, bOpsLabels, *bOpsValues, now)
		}
	}

	// bQuotaUsed
	for _, bucket := range stats.BucketsStats {
		bQuotaUsedLabels := pdata.NewStringMap()
		bQuotaUsedLabels.Insert(metadata.Labels.Buckets, bucket.Name)
		bQuotaUsedValues := bucket.BasicStats.QuotaPercentUsed
		if bQuotaUsedValues == nil {
			c.logger.Info(
				"failed to collect metric",
				zap.String("metric", "bQuotaUsed"),
			)
		} else {
			addToDoubleMetric(bQuotaUsed, bQuotaUsedLabels, *bQuotaUsedValues, now)
		}
	}

	// cmdGet
	cmdGetLabels := pdata.NewStringMap()
	cmdGetValues := stats.NodeStats.Nodes[0].InterestingStats.CmdGet
	if cmdGetValues == nil {
		c.logger.Info(
			"failed to collect metric",
			zap.String("metric", "cmdGet"),
		)
	} else {
		addToDoubleMetric(cmdGet, cmdGetLabels, *cmdGetValues, now)
	}

	// cpuUtilizationRate
	cpuUtilizationRateLabels := pdata.NewStringMap()
	cpuUtilizationRateValues := stats.NodeStats.Nodes[0].SystemStats.CPUUtilizationRate
	if cpuUtilizationRateValues == nil {
		c.logger.Info(
			"failed to collect metric",
			zap.String("metric", "cpuUtilizationRate"),
		)
	} else {
		addToDoubleMetric(cpuUtilizationRate, cpuUtilizationRateLabels, *cpuUtilizationRateValues, now)
	}

	// currItems
	currItemsLabels := pdata.NewStringMap()
	currItemsValues := stats.NodeStats.Nodes[0].InterestingStats.CurrItems
	if currItemsValues == nil {
		c.logger.Info(
			"failed to collect metric",
			zap.String("metric", "currItems"),
		)
	} else {
		addToIntMetric(currItems, currItemsLabels, *currItemsValues, now)
	}

	// currItemsTot
	currItemsTotLabels := pdata.NewStringMap()
	currItemsTotValues := stats.NodeStats.Nodes[0].InterestingStats.CurrItemsTot
	if currItemsTotValues == nil {
		c.logger.Info(
			"failed to collect metric",
			zap.String("metric", "currItemsTot"),
		)
	} else {
		addToIntMetric(currItemsTot, currItemsTotLabels, *currItemsTotValues, now)
	}

	// diskFetches
	diskFetchesLabels := pdata.NewStringMap()
	diskFetchesValues := stats.NodeStats.Nodes[0].InterestingStats.EpBgFetched
	if currItemsTotValues == nil {
		c.logger.Info(
			"failed to collect metric",
			zap.String("metric", "diskFetches"),
		)
	} else {
		addToDoubleMetric(diskFetches, diskFetchesLabels, *diskFetchesValues, now)
	}

	// getHits
	getHitsLabels := pdata.NewStringMap()
	getHitsValues := stats.NodeStats.Nodes[0].InterestingStats.GetHits
	if getHitsValues == nil {
		c.logger.Info(
			"failed to collect metric",
			zap.String("metric", "getHits"),
		)
	} else {
		addToDoubleMetric(getHits, getHitsLabels, *getHitsValues, now)
	}

	// memFree
	memFreeLabels := pdata.NewStringMap()
	memFreeValues := stats.NodeStats.Nodes[0].SystemStats.MemFree
	if memFreeValues == nil {
		c.logger.Info(
			"failed to collect metric",
			zap.String("metric", "memFree"),
		)
	} else {
		addToIntMetric(memFree, memFreeLabels, *memFreeValues, now)
	}

	// memTotal
	memTotalLabels := pdata.NewStringMap()
	memTotalValues := stats.NodeStats.Nodes[0].SystemStats.MemTotal
	if memTotalValues == nil {
		c.logger.Info(
			"failed to collect metric",
			zap.String("metric", "memTotal"),
		)
	} else {
		addToIntMetric(memTotal, memTotalLabels, *memTotalValues, now)
	}

	// memUsed
	memUsedLabels := pdata.NewStringMap()
	memUsedValues := stats.NodeStats.Nodes[0].InterestingStats.MemUsed
	if memUsedValues == nil {
		c.logger.Info(
			"failed to collect metric",
			zap.String("metric", "memUsed"),
		)
	} else {
		addToIntMetric(memUsed, memUsedLabels, *memUsedValues, now)
	}

	// ops
	opsLabels := pdata.NewStringMap()
	opsValues := stats.NodeStats.Nodes[0].InterestingStats.Ops
	if opsValues == nil {
		c.logger.Info(
			"failed to collect metric",
			zap.String("metric", "ops"),
		)
	} else {
		addToDoubleMetric(ops, opsLabels, *opsValues, now)
	}

	// swapTotal
	swapTotalLabels := pdata.NewStringMap()
	swapTotalValues := stats.NodeStats.Nodes[0].SystemStats.SwapTotal
	if swapTotalValues == nil {
		c.logger.Info(
			"failed to collect metric",
			zap.String("metric", "swapTotal"),
		)
	} else {
		addToIntMetric(swapTotal, swapTotalLabels, *swapTotalValues, now)
	}

	// swapUsed
	swapUsedLabels := pdata.NewStringMap()
	swapUsedValues := stats.NodeStats.Nodes[0].SystemStats.SwapUsed
	if swapUsedValues == nil {
		c.logger.Info(
			"failed to collect metric",
			zap.String("metric", "swapUsed"),
		)
	} else {
		addToIntMetric(swapUsed, swapUsedLabels, *swapUsedValues, now)
	}

	// uptime
	uptimeLabels := pdata.NewStringMap()
	uptimeValues, ok := c.parseInt("uptime", *stats.NodeStats.Nodes[0].Uptime)
	if !ok {
		c.logger.Info(
			"failed to collect metric",
			zap.String("metric", "swapUsed"),
		)
	} else {
		addToIntMetric(uptime, uptimeLabels, uptimeValues, now)
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

// parseInt converts string to int64.
func (p *couchbaseScraper) parseInt(key, value string) (int64, bool) {
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		p.logger.Info(
			"invalid value",
			zap.String("expectedType", "int"),
			zap.String("key", key),
			zap.String("value", value),
		)
		return 0, false
	}
	return i, true
}
