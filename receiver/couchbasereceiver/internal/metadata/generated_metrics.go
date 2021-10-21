// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/model/pdata"
)

// Type is the component type name.
const Type config.Type = "couchbasereceiver"

// MetricIntf is an interface to generically interact with generated metric.
type MetricIntf interface {
	Name() string
	New() pdata.Metric
	Init(metric pdata.Metric)
}

// Intentionally not exposing this so that it is opaque and can change freely.
type metricImpl struct {
	name     string
	initFunc func(pdata.Metric)
}

// Name returns the metric name.
func (m *metricImpl) Name() string {
	return m.name
}

// New creates a metric object preinitialized.
func (m *metricImpl) New() pdata.Metric {
	metric := pdata.NewMetric()
	m.Init(metric)
	return metric
}

// Init initializes the provided metric object.
func (m *metricImpl) Init(metric pdata.Metric) {
	m.initFunc(metric)
}

type metricStruct struct {
	CouchbaseBDataUsed          MetricIntf
	CouchbaseBDiskFetches       MetricIntf
	CouchbaseBDiskUsed          MetricIntf
	CouchbaseBItemCount         MetricIntf
	CouchbaseBMemUsed           MetricIntf
	CouchbaseBOps               MetricIntf
	CouchbaseBQuotaUsed         MetricIntf
	CouchbaseCmdGet             MetricIntf
	CouchbaseCPUUtilizationRate MetricIntf
	CouchbaseCurrItems          MetricIntf
	CouchbaseCurrItemsTot       MetricIntf
	CouchbaseDiskFetches        MetricIntf
	CouchbaseGetHits            MetricIntf
	CouchbaseMemFree            MetricIntf
	CouchbaseMemTotal           MetricIntf
	CouchbaseMemUsed            MetricIntf
	CouchbaseOps                MetricIntf
	CouchbaseSwapTotal          MetricIntf
	CouchbaseSwapUsed           MetricIntf
	CouchbaseUptime             MetricIntf
}

// Names returns a list of all the metric name strings.
func (m *metricStruct) Names() []string {
	return []string{
		"couchbase.b_data_used",
		"couchbase.b_disk_fetches",
		"couchbase.b_disk_used",
		"couchbase.b_item_count",
		"couchbase.b_mem_used",
		"couchbase.b_ops",
		"couchbase.b_quota_used",
		"couchbase.cmd_get",
		"couchbase.cpu_utilization_rate",
		"couchbase.curr_items",
		"couchbase.curr_items_tot",
		"couchbase.disk_fetches",
		"couchbase.get_hits",
		"couchbase.mem_free",
		"couchbase.mem_total",
		"couchbase.mem_used",
		"couchbase.ops",
		"couchbase.swap_total",
		"couchbase.swap_used",
		"couchbase.uptime",
	}
}

var metricsByName = map[string]MetricIntf{
	"couchbase.b_data_used":          Metrics.CouchbaseBDataUsed,
	"couchbase.b_disk_fetches":       Metrics.CouchbaseBDiskFetches,
	"couchbase.b_disk_used":          Metrics.CouchbaseBDiskUsed,
	"couchbase.b_item_count":         Metrics.CouchbaseBItemCount,
	"couchbase.b_mem_used":           Metrics.CouchbaseBMemUsed,
	"couchbase.b_ops":                Metrics.CouchbaseBOps,
	"couchbase.b_quota_used":         Metrics.CouchbaseBQuotaUsed,
	"couchbase.cmd_get":              Metrics.CouchbaseCmdGet,
	"couchbase.cpu_utilization_rate": Metrics.CouchbaseCPUUtilizationRate,
	"couchbase.curr_items":           Metrics.CouchbaseCurrItems,
	"couchbase.curr_items_tot":       Metrics.CouchbaseCurrItemsTot,
	"couchbase.disk_fetches":         Metrics.CouchbaseDiskFetches,
	"couchbase.get_hits":             Metrics.CouchbaseGetHits,
	"couchbase.mem_free":             Metrics.CouchbaseMemFree,
	"couchbase.mem_total":            Metrics.CouchbaseMemTotal,
	"couchbase.mem_used":             Metrics.CouchbaseMemUsed,
	"couchbase.ops":                  Metrics.CouchbaseOps,
	"couchbase.swap_total":           Metrics.CouchbaseSwapTotal,
	"couchbase.swap_used":            Metrics.CouchbaseSwapUsed,
	"couchbase.uptime":               Metrics.CouchbaseUptime,
}

func (m *metricStruct) ByName(n string) MetricIntf {
	return metricsByName[n]
}

// Metrics contains a set of methods for each metric that help with
// manipulating those metrics.
var Metrics = &metricStruct{
	&metricImpl{
		"couchbase.b_data_used",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.b_data_used")
			metric.SetDescription("Bytes in memory used by Index.")
			metric.SetUnit("by")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.b_disk_fetches",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.b_disk_fetches")
			metric.SetDescription("Number of reads per second from disk.")
			metric.SetUnit("number/sec")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.b_disk_used",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.b_disk_used")
			metric.SetDescription("Bytes on disk used by Index.")
			metric.SetUnit("by")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.b_item_count",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.b_item_count")
			metric.SetDescription("Number of items.")
			metric.SetUnit("number/sec")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.b_mem_used",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.b_mem_used")
			metric.SetDescription("Total memory used in bytes.")
			metric.SetUnit("by")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.b_ops",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.b_ops")
			metric.SetDescription("Number of operations per second.")
			metric.SetUnit("number/sec")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.b_quota_used",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.b_quota_used")
			metric.SetDescription("Percentage of index RAM quota.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.cmd_get",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.cmd_get")
			metric.SetDescription("Number of reads(get operations) per second.")
			metric.SetUnit("number/sec")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.cpu_utilization_rate",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.cpu_utilization_rate")
			metric.SetDescription("Percentage of CPU in use across all available cores.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.curr_items",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.curr_items")
			metric.SetDescription("Number of active items in all buckets.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.curr_items_tot",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.curr_items_tot")
			metric.SetDescription("Total number of items in all buckets.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.disk_fetches",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.disk_fetches")
			metric.SetDescription("Number of reads per second from disk.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.get_hits",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.get_hits")
			metric.SetDescription("Number of reads(get operations) from RAM per second.")
			metric.SetUnit("number/sec")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.mem_free",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.mem_free")
			metric.SetDescription("Bytes of RAM available.")
			metric.SetUnit("by")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.mem_total",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.mem_total")
			metric.SetDescription("Total bytes capacity of RAM.")
			metric.SetUnit("by")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.mem_used",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.mem_used")
			metric.SetDescription("Total memory used in bytes.")
			metric.SetUnit("by")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.ops",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.ops")
			metric.SetDescription("Number of operations per second.")
			metric.SetUnit("number/sec")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.swap_total",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.swap_total")
			metric.SetDescription("Total bytes capacity.")
			metric.SetUnit("by")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.swap_used",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.swap_used")
			metric.SetDescription("Bytes of swap space in use.")
			metric.SetUnit("by")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.uptime",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.uptime")
			metric.SetDescription("Total time the couchbase server has been running.")
			metric.SetUnit("s")
			metric.SetDataType(pdata.MetricDataTypeSum)
			metric.Sum().SetIsMonotonic(true)
			metric.Sum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
}

// M contains a set of methods for each metric that help with
// manipulating those metrics. M is an alias for Metrics
var M = Metrics

// Labels contains the possible metric labels that can be used.
var Labels = struct {
	// Buckets (The name of the bucket.)
	Buckets string
}{
	"buckets",
}

// L contains the possible metric labels that can be used. L is an alias for
// Labels.
var L = Labels
