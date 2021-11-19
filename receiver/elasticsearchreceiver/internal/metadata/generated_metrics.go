// Code generated by mdatagen. DO NOT EDIT.

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
	ElasticsearchCacheMemoryUsage    MetricIntf
	ElasticsearchCurrentDocuments    MetricIntf
	ElasticsearchDataNodes           MetricIntf
	ElasticsearchEvictions           MetricIntf
	ElasticsearchGcCollection        MetricIntf
	ElasticsearchGcCollectionTime    MetricIntf
	ElasticsearchHTTPConnections     MetricIntf
	ElasticsearchMemoryUsage         MetricIntf
	ElasticsearchNetwork             MetricIntf
	ElasticsearchNodes               MetricIntf
	ElasticsearchOpenFiles           MetricIntf
	ElasticsearchOperationTime       MetricIntf
	ElasticsearchOperations          MetricIntf
	ElasticsearchPeakThreads         MetricIntf
	ElasticsearchServerConnections   MetricIntf
	ElasticsearchShards              MetricIntf
	ElasticsearchStorageSize         MetricIntf
	ElasticsearchThreadPoolActive    MetricIntf
	ElasticsearchThreadPoolCompleted MetricIntf
	ElasticsearchThreadPoolQueue     MetricIntf
	ElasticsearchThreadPoolRejected  MetricIntf
	ElasticsearchThreadPoolThreads   MetricIntf
	ElasticsearchThreads             MetricIntf
}

// Names returns a list of all the metric name strings.
func (m *metricStruct) Names() []string {
	return []string{
		"elasticsearch.cache_memory_usage",
		"elasticsearch.current_documents",
		"elasticsearch.data_nodes",
		"elasticsearch.evictions",
		"elasticsearch.gc_collection",
		"elasticsearch.gc_collection_time",
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
		"elasticsearch.thread_pool.active",
		"elasticsearch.thread_pool.completed",
		"elasticsearch.thread_pool.queue",
		"elasticsearch.thread_pool.rejected",
		"elasticsearch.thread_pool.threads",
		"elasticsearch.threads",
	}
}

var metricsByName = map[string]MetricIntf{
	"elasticsearch.cache_memory_usage":    Metrics.ElasticsearchCacheMemoryUsage,
	"elasticsearch.current_documents":     Metrics.ElasticsearchCurrentDocuments,
	"elasticsearch.data_nodes":            Metrics.ElasticsearchDataNodes,
	"elasticsearch.evictions":             Metrics.ElasticsearchEvictions,
	"elasticsearch.gc_collection":         Metrics.ElasticsearchGcCollection,
	"elasticsearch.gc_collection_time":    Metrics.ElasticsearchGcCollectionTime,
	"elasticsearch.http_connections":      Metrics.ElasticsearchHTTPConnections,
	"elasticsearch.memory_usage":          Metrics.ElasticsearchMemoryUsage,
	"elasticsearch.network":               Metrics.ElasticsearchNetwork,
	"elasticsearch.nodes":                 Metrics.ElasticsearchNodes,
	"elasticsearch.open_files":            Metrics.ElasticsearchOpenFiles,
	"elasticsearch.operation_time":        Metrics.ElasticsearchOperationTime,
	"elasticsearch.operations":            Metrics.ElasticsearchOperations,
	"elasticsearch.peak_threads":          Metrics.ElasticsearchPeakThreads,
	"elasticsearch.server_connections":    Metrics.ElasticsearchServerConnections,
	"elasticsearch.shards":                Metrics.ElasticsearchShards,
	"elasticsearch.storage_size":          Metrics.ElasticsearchStorageSize,
	"elasticsearch.thread_pool.active":    Metrics.ElasticsearchThreadPoolActive,
	"elasticsearch.thread_pool.completed": Metrics.ElasticsearchThreadPoolCompleted,
	"elasticsearch.thread_pool.queue":     Metrics.ElasticsearchThreadPoolQueue,
	"elasticsearch.thread_pool.rejected":  Metrics.ElasticsearchThreadPoolRejected,
	"elasticsearch.thread_pool.threads":   Metrics.ElasticsearchThreadPoolThreads,
	"elasticsearch.threads":               Metrics.ElasticsearchThreads,
}

func (m *metricStruct) ByName(n string) MetricIntf {
	return metricsByName[n]
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
			metric.SetDataType(pdata.MetricDataTypeSum)
			metric.Sum().SetIsMonotonic(true)
			metric.Sum().SetAggregationTemporality(pdata.MetricAggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"elasticsearch.gc_collection",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.gc_collection")
			metric.SetDescription("Garbage collection count.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeSum)
			metric.Sum().SetIsMonotonic(true)
			metric.Sum().SetAggregationTemporality(pdata.MetricAggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"elasticsearch.gc_collection_time",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.gc_collection_time")
			metric.SetDescription("Garbage collection time.")
			metric.SetUnit("ms")
			metric.SetDataType(pdata.MetricDataTypeSum)
			metric.Sum().SetIsMonotonic(true)
			metric.Sum().SetAggregationTemporality(pdata.MetricAggregationTemporalityCumulative)
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
			metric.SetDataType(pdata.MetricDataTypeSum)
			metric.Sum().SetIsMonotonic(true)
			metric.Sum().SetAggregationTemporality(pdata.MetricAggregationTemporalityCumulative)
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
			metric.SetDataType(pdata.MetricDataTypeSum)
			metric.Sum().SetIsMonotonic(true)
			metric.Sum().SetAggregationTemporality(pdata.MetricAggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"elasticsearch.operations",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.operations")
			metric.SetDescription("Number of operations completed")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeSum)
			metric.Sum().SetIsMonotonic(true)
			metric.Sum().SetAggregationTemporality(pdata.MetricAggregationTemporalityCumulative)
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
		"elasticsearch.thread_pool.active",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.thread_pool.active")
			metric.SetDescription("Number of active threads in the thread pool.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"elasticsearch.thread_pool.completed",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.thread_pool.completed")
			metric.SetDescription("Number of tasks completed by the thread pool executor.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeSum)
			metric.Sum().SetIsMonotonic(false)
			metric.Sum().SetAggregationTemporality(pdata.MetricAggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"elasticsearch.thread_pool.queue",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.thread_pool.queue")
			metric.SetDescription("Number of tasks in the queue for the thread pool.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"elasticsearch.thread_pool.rejected",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.thread_pool.rejected")
			metric.SetDescription("Number of tasks rejected by the thread pool executor.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeSum)
			metric.Sum().SetIsMonotonic(false)
			metric.Sum().SetAggregationTemporality(pdata.MetricAggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"elasticsearch.thread_pool.threads",
		func(metric pdata.Metric) {
			metric.SetName("elasticsearch.thread_pool.threads")
			metric.SetDescription("Number of threads in the pool.")
			metric.SetUnit("1")
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

// Attributes contains the possible metric attributes that can be used.
var Attributes = struct {
	// CacheName (Type of cache)
	CacheName string
	// Direction (Data direction)
	Direction string
	// DocumentType (Type of document count)
	DocumentType string
	// GcType (Type of garbage collection)
	GcType string
	// MemoryType (Type of memory)
	MemoryType string
	// Operation (Type of operation)
	Operation string
	// ServerName (The name of the server or node the metric is based on.)
	ServerName string
	// ShardType (State of the shard)
	ShardType string
	// ThreadPoolName (Thread pool name)
	ThreadPoolName string
}{
	"cache_name",
	"direction",
	"document_type",
	"gc_type",
	"memory_type",
	"operation",
	"server_name",
	"shard_type",
	"thread_pool_name",
}

// A is an alias for Attributes.
var A = Attributes

// AttributeCacheName are the possible values that the attribute "cache_name" can have.
var AttributeCacheName = struct {
	Field   string
	Query   string
	Request string
}{
	"field",
	"query",
	"request",
}

// AttributeDirection are the possible values that the attribute "direction" can have.
var AttributeDirection = struct {
	Receive  string
	Transmit string
}{
	"receive",
	"transmit",
}

// AttributeDocumentType are the possible values that the attribute "document_type" can have.
var AttributeDocumentType = struct {
	Live    string
	Deleted string
}{
	"live",
	"deleted",
}

// AttributeGcType are the possible values that the attribute "gc_type" can have.
var AttributeGcType = struct {
	Young string
	Old   string
}{
	"young",
	"old",
}

// AttributeMemoryType are the possible values that the attribute "memory_type" can have.
var AttributeMemoryType = struct {
	Heap    string
	NonHeap string
}{
	"heap",
	"non-heap",
}

// AttributeOperation are the possible values that the attribute "operation" can have.
var AttributeOperation = struct {
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

// AttributeShardType are the possible values that the attribute "shard_type" can have.
var AttributeShardType = struct {
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
