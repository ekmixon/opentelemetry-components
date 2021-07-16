// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/model/pdata"
)

// Type is the component type name.
const Type config.Type = "mysqlreceiver"

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
	MysqlBufferPoolOperations MetricIntf
	MysqlBufferPoolPages      MetricIntf
	MysqlBufferPoolSize       MetricIntf
	MysqlCommands             MetricIntf
	MysqlDoubleWrites         MetricIntf
	MysqlHandlers             MetricIntf
	MysqlLocks                MetricIntf
	MysqlLogOperations        MetricIntf
	MysqlOperations           MetricIntf
	MysqlPageOperations       MetricIntf
	MysqlRowLocks             MetricIntf
	MysqlRowOperations        MetricIntf
	MysqlSorts                MetricIntf
	MysqlThreads              MetricIntf
}

// Names returns a list of all the metric name strings.
func (m *metricStruct) Names() []string {
	return []string{
		"mysql.buffer_pool_operations",
		"mysql.buffer_pool_pages",
		"mysql.buffer_pool_size",
		"mysql.commands",
		"mysql.double_writes",
		"mysql.handlers",
		"mysql.locks",
		"mysql.log_operations",
		"mysql.operations",
		"mysql.page_operations",
		"mysql.row_locks",
		"mysql.row_operations",
		"mysql.sorts",
		"mysql.threads",
	}
}

var metricsByName = map[string]MetricIntf{
	"mysql.buffer_pool_operations": Metrics.MysqlBufferPoolOperations,
	"mysql.buffer_pool_pages":      Metrics.MysqlBufferPoolPages,
	"mysql.buffer_pool_size":       Metrics.MysqlBufferPoolSize,
	"mysql.commands":               Metrics.MysqlCommands,
	"mysql.double_writes":          Metrics.MysqlDoubleWrites,
	"mysql.handlers":               Metrics.MysqlHandlers,
	"mysql.locks":                  Metrics.MysqlLocks,
	"mysql.log_operations":         Metrics.MysqlLogOperations,
	"mysql.operations":             Metrics.MysqlOperations,
	"mysql.page_operations":        Metrics.MysqlPageOperations,
	"mysql.row_locks":              Metrics.MysqlRowLocks,
	"mysql.row_operations":         Metrics.MysqlRowOperations,
	"mysql.sorts":                  Metrics.MysqlSorts,
	"mysql.threads":                Metrics.MysqlThreads,
}

func (m *metricStruct) ByName(n string) MetricIntf {
	return metricsByName[n]
}

func (m *metricStruct) FactoriesByName() map[string]func(pdata.Metric) {
	return map[string]func(pdata.Metric){
		Metrics.MysqlBufferPoolOperations.Name(): Metrics.MysqlBufferPoolOperations.Init,
		Metrics.MysqlBufferPoolPages.Name():      Metrics.MysqlBufferPoolPages.Init,
		Metrics.MysqlBufferPoolSize.Name():       Metrics.MysqlBufferPoolSize.Init,
		Metrics.MysqlCommands.Name():             Metrics.MysqlCommands.Init,
		Metrics.MysqlDoubleWrites.Name():         Metrics.MysqlDoubleWrites.Init,
		Metrics.MysqlHandlers.Name():             Metrics.MysqlHandlers.Init,
		Metrics.MysqlLocks.Name():                Metrics.MysqlLocks.Init,
		Metrics.MysqlLogOperations.Name():        Metrics.MysqlLogOperations.Init,
		Metrics.MysqlOperations.Name():           Metrics.MysqlOperations.Init,
		Metrics.MysqlPageOperations.Name():       Metrics.MysqlPageOperations.Init,
		Metrics.MysqlRowLocks.Name():             Metrics.MysqlRowLocks.Init,
		Metrics.MysqlRowOperations.Name():        Metrics.MysqlRowOperations.Init,
		Metrics.MysqlSorts.Name():                Metrics.MysqlSorts.Init,
		Metrics.MysqlThreads.Name():              Metrics.MysqlThreads.Init,
	}
}

// Metrics contains a set of methods for each metric that help with
// manipulating those metrics.
var Metrics = &metricStruct{
	&metricImpl{
		"mysql.buffer_pool_operations",
		func(metric pdata.Metric) {
			metric.SetName("mysql.buffer_pool_operations")
			metric.SetDescription("Buffer pool operation count")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"mysql.buffer_pool_pages",
		func(metric pdata.Metric) {
			metric.SetName("mysql.buffer_pool_pages")
			metric.SetDescription("Buffer pool page count")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"mysql.buffer_pool_size",
		func(metric pdata.Metric) {
			metric.SetName("mysql.buffer_pool_size")
			metric.SetDescription("Buffer pool size")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"mysql.commands",
		func(metric pdata.Metric) {
			metric.SetName("mysql.commands")
			metric.SetDescription("MySQL command count")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"mysql.double_writes",
		func(metric pdata.Metric) {
			metric.SetName("mysql.double_writes")
			metric.SetDescription("InnoDB doublewrite buffer count")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"mysql.handlers",
		func(metric pdata.Metric) {
			metric.SetName("mysql.handlers")
			metric.SetDescription("MySQL handler count")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"mysql.locks",
		func(metric pdata.Metric) {
			metric.SetName("mysql.locks")
			metric.SetDescription("MySQL lock count")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"mysql.log_operations",
		func(metric pdata.Metric) {
			metric.SetName("mysql.log_operations")
			metric.SetDescription("InndoDB log operation count")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"mysql.operations",
		func(metric pdata.Metric) {
			metric.SetName("mysql.operations")
			metric.SetDescription("InndoDB operation count")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"mysql.page_operations",
		func(metric pdata.Metric) {
			metric.SetName("mysql.page_operations")
			metric.SetDescription("InndoDB page operation count")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"mysql.row_locks",
		func(metric pdata.Metric) {
			metric.SetName("mysql.row_locks")
			metric.SetDescription("InndoDB row lock count")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"mysql.row_operations",
		func(metric pdata.Metric) {
			metric.SetName("mysql.row_operations")
			metric.SetDescription("InndoDB row operation count")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"mysql.sorts",
		func(metric pdata.Metric) {
			metric.SetName("mysql.sorts")
			metric.SetDescription("MySQL sort count")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"mysql.threads",
		func(metric pdata.Metric) {
			metric.SetName("mysql.threads")
			metric.SetDescription("Thread count")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
}

// M contains a set of methods for each metric that help with
// manipulating those metrics. M is an alias for Metrics
var M = Metrics

// Labels contains the possible metric labels that can be used.
var Labels = struct {
	// BufferPoolOperationsState (The buffer pool operations types)
	BufferPoolOperationsState string
	// BufferPoolPagesState (The buffer pool pages types)
	BufferPoolPagesState string
	// BufferPoolSizeState (The buffer pool size types)
	BufferPoolSizeState string
	// CommandState (The command types)
	CommandState string
	// DoubleWritesState (The doublewrite types)
	DoubleWritesState string
	// HandlerState (The handler types)
	HandlerState string
	// LocksState (The table locks type)
	LocksState string
	// LogOperationsState (The log operation types)
	LogOperationsState string
	// OperationsState (The operation types)
	OperationsState string
	// PageOperationsState (The page operation types)
	PageOperationsState string
	// RowLocksState (The row lock type)
	RowLocksState string
	// RowOperationsState (The row operation type)
	RowOperationsState string
	// SortsState (The sort count type)
	SortsState string
	// ThreadsState (The thread count type)
	ThreadsState string
}{
	"operation",
	"kind",
	"kind",
	"command",
	"kind",
	"kind",
	"kind",
	"operation",
	"operation",
	"operation",
	"kind",
	"operation",
	"kind",
	"kind",
}

// L contains the possible metric labels that can be used. L is an alias for
// Labels.
var L = Labels

// LabelBufferPoolOperationsState are the possible values that the label "buffer_pool_operations_state" can have.
var LabelBufferPoolOperationsState = struct {
	ReadAheadRnd     string
	ReadAhead        string
	ReadAheadEvicted string
	ReadRequests     string
	Reads            string
	WaitFree         string
	WriteRequests    string
}{
	"read_ahead_rnd",
	"read_ahead",
	"read_ahead_evicted",
	"read_requests",
	"reads",
	"wait_free",
	"write_requests",
}

// LabelBufferPoolPagesState are the possible values that the label "buffer_pool_pages_state" can have.
var LabelBufferPoolPagesState = struct {
	Data    string
	Dirty   string
	Flushed string
	Free    string
	Misc    string
	Total   string
}{
	"data",
	"dirty",
	"flushed",
	"free",
	"misc",
	"total",
}

// LabelBufferPoolSizeState are the possible values that the label "buffer_pool_size_state" can have.
var LabelBufferPoolSizeState = struct {
	Data  string
	Dirty string
	Size  string
}{
	"data",
	"dirty",
	"size",
}

// LabelCommandState are the possible values that the label "command_state" can have.
var LabelCommandState = struct {
	Execute      string
	Close        string
	Fetch        string
	Prepare      string
	Reset        string
	SendLongData string
}{
	"execute",
	"close",
	"fetch",
	"prepare",
	"reset",
	"send_long_data",
}

// LabelDoubleWritesState are the possible values that the label "double_writes_state" can have.
var LabelDoubleWritesState = struct {
	PagesWritten string
	Writes       string
}{
	"pages_written",
	"writes",
}

// LabelHandlerState are the possible values that the label "handler_state" can have.
var LabelHandlerState = struct {
	Commit            string
	Delete            string
	Discover          string
	ExternalLock      string
	MrrInit           string
	Prepare           string
	ReadFirst         string
	ReadKey           string
	ReadLast          string
	ReadNext          string
	ReadPrev          string
	ReadRnd           string
	ReadRndNext       string
	Rollback          string
	Savepoint         string
	SavepointRollback string
	Update            string
	Write             string
}{
	"commit",
	"delete",
	"discover",
	"external_lock",
	"mrr_init",
	"prepare",
	"read_first",
	"read_key",
	"read_last",
	"read_next",
	"read_prev",
	"read_rnd",
	"read_rnd_next",
	"rollback",
	"savepoint",
	"savepoint_rollback",
	"update",
	"write",
}

// LabelLocksState are the possible values that the label "locks_state" can have.
var LabelLocksState = struct {
	Immediate string
	Waited    string
}{
	"immediate",
	"waited",
}

// LabelLogOperationsState are the possible values that the label "log_operations_state" can have.
var LabelLogOperationsState = struct {
	Waits         string
	WriteRequests string
	Writes        string
}{
	"waits",
	"write_requests",
	"writes",
}

// LabelOperationsState are the possible values that the label "operations_state" can have.
var LabelOperationsState = struct {
	Fsyncs string
	Reads  string
	Writes string
}{
	"fsyncs",
	"reads",
	"writes",
}

// LabelPageOperationsState are the possible values that the label "page_operations_state" can have.
var LabelPageOperationsState = struct {
	Created string
	Read    string
	Written string
}{
	"created",
	"read",
	"written",
}

// LabelRowLocksState are the possible values that the label "row_locks_state" can have.
var LabelRowLocksState = struct {
	Waits string
	Time  string
}{
	"waits",
	"time",
}

// LabelRowOperationsState are the possible values that the label "row_operations_state" can have.
var LabelRowOperationsState = struct {
	Deleted  string
	Inserted string
	Read     string
	Updated  string
}{
	"deleted",
	"inserted",
	"read",
	"updated",
}

// LabelSortsState are the possible values that the label "sorts_state" can have.
var LabelSortsState = struct {
	MergePasses string
	Range       string
	Rows        string
	Scan        string
}{
	"merge_passes",
	"range",
	"rows",
	"scan",
}

// LabelThreadsState are the possible values that the label "threads_state" can have.
var LabelThreadsState = struct {
	Cached    string
	Connected string
	Created   string
	Running   string
}{
	"cached",
	"connected",
	"created",
	"running",
}
