// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/model/pdata"
)

// Type is the component type name.
const Type config.Type = "postgresql"

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
	PostgresqlBackends   MetricIntf
	PostgresqlBlocksRead MetricIntf
	PostgresqlCommits    MetricIntf
	PostgresqlDbSize     MetricIntf
	PostgresqlOperations MetricIntf
	PostgresqlRollbacks  MetricIntf
	PostgresqlRows       MetricIntf
}

// Names returns a list of all the metric name strings.
func (m *metricStruct) Names() []string {
	return []string{
		"postgresql.backends",
		"postgresql.blocks_read",
		"postgresql.commits",
		"postgresql.db_size",
		"postgresql.operations",
		"postgresql.rollbacks",
		"postgresql.rows",
	}
}

var metricsByName = map[string]MetricIntf{
	"postgresql.backends":    Metrics.PostgresqlBackends,
	"postgresql.blocks_read": Metrics.PostgresqlBlocksRead,
	"postgresql.commits":     Metrics.PostgresqlCommits,
	"postgresql.db_size":     Metrics.PostgresqlDbSize,
	"postgresql.operations":  Metrics.PostgresqlOperations,
	"postgresql.rollbacks":   Metrics.PostgresqlRollbacks,
	"postgresql.rows":        Metrics.PostgresqlRows,
}

func (m *metricStruct) ByName(n string) MetricIntf {
	return metricsByName[n]
}

// Metrics contains a set of methods for each metric that help with
// manipulating those metrics.
var Metrics = &metricStruct{
	&metricImpl{
		"postgresql.backends",
		func(metric pdata.Metric) {
			metric.SetName("postgresql.backends")
			metric.SetDescription("The number of backends.")
			metric.SetUnit("")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"postgresql.blocks_read",
		func(metric pdata.Metric) {
			metric.SetName("postgresql.blocks_read")
			metric.SetDescription("The number of blocks read.")
			metric.SetUnit("")
			metric.SetDataType(pdata.MetricDataTypeSum)
			metric.Sum().SetIsMonotonic(true)
			metric.Sum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"postgresql.commits",
		func(metric pdata.Metric) {
			metric.SetName("postgresql.commits")
			metric.SetDescription("The number of commits.")
			metric.SetUnit("")
			metric.SetDataType(pdata.MetricDataTypeSum)
			metric.Sum().SetIsMonotonic(true)
			metric.Sum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"postgresql.db_size",
		func(metric pdata.Metric) {
			metric.SetName("postgresql.db_size")
			metric.SetDescription("The database disk usage.")
			metric.SetUnit("")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"postgresql.operations",
		func(metric pdata.Metric) {
			metric.SetName("postgresql.operations")
			metric.SetDescription("The number of db row operations.")
			metric.SetUnit("")
			metric.SetDataType(pdata.MetricDataTypeSum)
			metric.Sum().SetIsMonotonic(true)
			metric.Sum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"postgresql.rollbacks",
		func(metric pdata.Metric) {
			metric.SetName("postgresql.rollbacks")
			metric.SetDescription("The number of rollbacks.")
			metric.SetUnit("")
			metric.SetDataType(pdata.MetricDataTypeSum)
			metric.Sum().SetIsMonotonic(true)
			metric.Sum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"postgresql.rows",
		func(metric pdata.Metric) {
			metric.SetName("postgresql.rows")
			metric.SetDescription("The number of rows in the database.")
			metric.SetUnit("")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
}

// M contains a set of methods for each metric that help with
// manipulating those metrics. M is an alias for Metrics
var M = Metrics

// Labels contains the possible metric labels that can be used.
var Labels = struct {
	// Database (The name of the database.)
	Database string
	// Operation (The database operation.)
	Operation string
	// Source (The block read source type.)
	Source string
	// State (The tuple (row) state.)
	State string
	// Table (The schema name followed by the table name.)
	Table string
}{
	"database",
	"operation",
	"source",
	"state",
	"table",
}

// L contains the possible metric labels that can be used. L is an alias for
// Labels.
var L = Labels

// LabelOperation are the possible values that the label "operation" can have.
var LabelOperation = struct {
	Ins    string
	Upd    string
	Del    string
	HotUpd string
}{
	"ins",
	"upd",
	"del",
	"hot_upd",
}

// LabelSource are the possible values that the label "source" can have.
var LabelSource = struct {
	HeapRead  string
	HeapHit   string
	IdxRead   string
	IdxHit    string
	ToastRead string
	ToastHit  string
	TidxRead  string
	TidxHit   string
}{
	"heap_read",
	"heap_hit",
	"idx_read",
	"idx_hit",
	"toast_read",
	"toast_hit",
	"tidx_read",
	"tidx_hit",
}

// LabelState are the possible values that the label "state" can have.
var LabelState = struct {
	Dead string
	Live string
}{
	"dead",
	"live",
}
