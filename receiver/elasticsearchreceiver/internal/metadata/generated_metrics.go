package metadata

import (
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/model/pdata"
)

// Type is the component type name.
const Type config.Type = "elasticsearchreceiver"

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
	ElasticsearchCacheMemoryUsage  MetricIntf
	ElasticsearchCurrentDocuments  MetricIntf
	ElasticsearchDataNodes         MetricIntf
	ElasticsearchEvictions         MetricIntf
	ElasticsearchGcCollection      MetricIntf
	ElasticsearchHTTPConnections   MetricIntf
	ElasticsearchMemoryUsage       MetricIntf
	ElasticsearchNetwork           MetricIntf
	ElasticsearchNodes             MetricIntf
	ElasticsearchOpenFiles         MetricIntf
	ElasticsearchOperationTime     MetricIntf
	ElasticsearchOperations        MetricIntf
	ElasticsearchPeakThreads       MetricIntf
	ElasticsearchServerConnections MetricIntf
	ElasticsearchShards            MetricIntf
	ElasticsearchStorageSize       MetricIntf
	ElasticsearchThreads           MetricIntf
}

// Names returns a list of all the metric name strings.
func (m *metricStruct) Names() []string {
	return []string{
		"elasticsearch.cache_memory_usage",
		"elasticsearch.current_documents",
		"elasticsearch.data_nodes",
		"elasticsearch.evictions",
		"elasticsearch.gc_collection",
		"elasticsearch.http_connections",
		"elasticsearch.memory_usage",
		"elasticsearch.network",
		"elasticsearch.nodes",
		"elasticsearch.open_files",
		"elasticsearch.operation_time",
		"elasticsearch.operations",
		"elasticsearch.peak_threads",
		"elasticsearch.server_connections",
		"elasticsearch.shards",
		"elasticsearch.storage_size",
		"elasticsearch.threads",
	}
}

var metricsByName = map[string]MetricIntf{
	"elasticsearch.cache_memory_usage": Metrics.ElasticsearchCacheMemoryUsage,
	"elasticsearch.current_documents":  Metrics.ElasticsearchCurrentDocuments,
	"elasticsearch.data_nodes":         Metrics.ElasticsearchDataNodes,
	"elasticsearch.evictions":          Metrics.ElasticsearchEvictions,
	"elasticsearch.gc_collection":      Metrics.ElasticsearchGcCollection,
	"elasticsearch.http_connections":   Metrics.ElasticsearchHTTPConnections,
	"elasticsearch.memory_usage":       Metrics.ElasticsearchMemoryUsage,
	"elasticsearch.network":            Metrics.ElasticsearchNetwork,
	"elasticsearch.nodes":              Metrics.ElasticsearchNodes,
	"elasticsearch.open_files":         Metrics.ElasticsearchOpenFiles,
	"elasticsearch.operation_time":     Metrics.ElasticsearchOperationTime,
	"elasticsearch.operations":         Metrics.ElasticsearchOperations,
	"elasticsearch.peak_threads":       Metrics.ElasticsearchPeakThreads,
	"elasticsearch.server_connections": Metrics.ElasticsearchServerConnections,
	"elasticsearch.shards":             Metrics.ElasticsearchShards,
	"elasticsearch.storage_size":       Metrics.ElasticsearchStorageSize,
	"elasticsearch.threads":            Metrics.ElasticsearchThreads,
}

func (m *metricStruct) ByName(n string) MetricIntf {
	return metricsByName[n]
}

func (m *metricStruct) FactoriesByName() map[string]func(pdata.Metric) {
	return map[string]func(pdata.Metric){
		Metrics.ElasticsearchCacheMemoryUsage.Name():  Metrics.ElasticsearchCacheMemoryUsage.Init,
		Metrics.ElasticsearchCurrentDocuments.Name():  Metrics.ElasticsearchCurrentDocuments.Init,
		Metrics.ElasticsearchDataNodes.Name():         Metrics.ElasticsearchDataNodes.Init,
		Metrics.ElasticsearchEvictions.Name():         Metrics.ElasticsearchEvictions.Init,
		Metrics.ElasticsearchGcCollection.Name():      Metrics.ElasticsearchGcCollection.Init,
		Metrics.ElasticsearchHTTPConnections.Name():   Metrics.ElasticsearchHTTPConnections.Init,
		Metrics.ElasticsearchMemoryUsage.Name():       Metrics.ElasticsearchMemoryUsage.Init,
		Metrics.ElasticsearchNetwork.Name():           Metrics.ElasticsearchNetwork.Init,
		Metrics.ElasticsearchNodes.Name():             Metrics.ElasticsearchNodes.Init,
		Metrics.ElasticsearchOpenFiles.Name():         Metrics.ElasticsearchOpenFiles.Init,
		Metrics.ElasticsearchOperationTime.Name():     Metrics.ElasticsearchOperationTime.Init,
		Metrics.ElasticsearchOperations.Name():        Metrics.ElasticsearchOperations.Init,
		Metrics.ElasticsearchPeakThreads.Name():       Metrics.ElasticsearchPeakThreads.Init,
		Metrics.ElasticsearchServerConnections.Name(): Metrics.ElasticsearchServerConnections.Init,
		Metrics.ElasticsearchShards.Name():            Metrics.ElasticsearchShards.Init,
		Metrics.ElasticsearchStorageSize.Name():       Metrics.ElasticsearchStorageSize.Init,
		Metrics.ElasticsearchThreads.Name():           Metrics.ElasticsearchThreads.Init,
	}
}

// Metrics contains a set of methods for each metric that help with
// manipulating those metrics.
var Metrics = &metricStruct{
	&metricImpl{
		"elasticsearch.cache_memory_usage",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.cache_memory_usage")
			metric.SetDescription("Size in bytes of the caches.")
			metric.SetUnit("by")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"elasticsearch.current_documents",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.current_documents")
			metric.SetDescription("Number of documents in the indexes on this node.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"elasticsearch.data_nodes",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.data_nodes")
			metric.SetDescription("Number of data nodes in the cluster.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"elasticsearch.evictions",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.evictions")
			metric.SetDescription("Evictions from each cache")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"elasticsearch.gc_collection",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.gc_collection")
			metric.SetDescription("Garbage collection count.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"elasticsearch.http_connections",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.http_connections")
			metric.SetDescription("Number of open HTTP connections to this node.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"elasticsearch.memory_usage",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.memory_usage")
			metric.SetDescription("Size in bytes of memory.")
			metric.SetUnit("by")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"elasticsearch.network",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.network")
			metric.SetDescription("Number of bytes transmitted and received on the network.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"elasticsearch.nodes",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.nodes")
			metric.SetDescription("Number of nodes in the cluster.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"elasticsearch.open_files",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.open_files")
			metric.SetDescription("Number of open file descriptors held by the server process.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"elasticsearch.operation_time",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.operation_time")
			metric.SetDescription("Time in ms spent on operations")
			metric.SetUnit("ms")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"elasticsearch.operations",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.operations")
			metric.SetDescription("Number of operations completed")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"elasticsearch.peak_threads",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.peak_threads")
			metric.SetDescription("Maximum number of open threads that have been open concurrently in the server JVM process.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"elasticsearch.server_connections",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.server_connections")
			metric.SetDescription("Number of open network connections to the server.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"elasticsearch.shards",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.shards")
			metric.SetDescription("Number of shards")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"elasticsearch.storage_size",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.storage_size")
			metric.SetDescription("Size in bytes of the document storage on this node.")
			metric.SetUnit("by")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"elasticsearch.threads",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.threads")
			metric.SetDescription("Number of open threads in the server JVM process.")
			metric.SetUnit("by")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
}

// M contains a set of methods for each metric that help with
// manipulating those metrics. M is an alias for Metrics
var M = Metrics

// Labels contains the possible metric labels that can be used.
var Labels = struct {
	// CacheName (type of cache)
	CacheName string
	// Direction (Data direction)
	Direction string
	// DocumentType (Document count type.)
	DocumentType string
	// GcType (type of garbage collection)
	GcType string
	// MemoryType (type of memory)
	MemoryType string
	// Operation (the operation_type)
	Operation string
	// ServerName (The name of the server or node the metric is based on.)
	ServerName string
	// ShardType (State of the shard)
	ShardType string
}{
	"cache_name",
	"direction",
	"document_type",
	"gc_type",
	"memory_type",
	"operation",
	"server_name",
	"shard_type",
}

// L contains the possible metric labels that can be used. L is an alias for
// Labels.
var L = Labels

// LabelCacheName are the possible values that the label "cache_name" can have.
var LabelCacheName = struct {
	Field   string
	Query   string
	Request string
}{
	"field",
	"query",
	"request",
}

// LabelDirection are the possible values that the label "direction" can have.
var LabelDirection = struct {
	Receive  string
	Transmit string
}{
	"receive",
	"transmit",
}

// LabelDocumentType are the possible values that the label "document_type" can have.
var LabelDocumentType = struct {
	Live    string
	Deleted string
}{
	"live",
	"deleted",
}

// LabelGcType are the possible values that the label "gc_type" can have.
var LabelGcType = struct {
	Young string
	Old   string
}{
	"young",
	"old",
}

// LabelMemoryType are the possible values that the label "memory_type" can have.
var LabelMemoryType = struct {
	Heap    string
	NonHeap string
}{
	"heap",
	"non-heap",
}

// LabelOperation are the possible values that the label "operation" can have.
var LabelOperation = struct {
	Get    string
	Delete string
	Index  string
	Query  string
	Fetch  string
}{
	"get",
	"delete",
	"index",
	"query",
	"fetch",
}

// LabelShardType are the possible values that the label "shard_type" can have.
var LabelShardType = struct {
	Active       string
	Relocating   string
	Initializing string
	Unassigned   string
}{
	"active",
	"relocating",
	"initializing",
	"unassigned",
}
