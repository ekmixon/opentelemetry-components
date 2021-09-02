package couchbasereceiver

import "github.com/observiq/opentelemetry-components/receiver/couchbasereceiver/internal/metadata"

type Metric struct {
	Name          string              // metric name, i.e. kv_collection_data_size_bytes
	Timestamp     int                 // resulting last timestamp value collected
	Value         string              // resulting last value collected
	Labels        []string            // metadata labels
	ValueType     string              // float or int64
	MetadataLabel metadata.MetricIntf // the metatdata generated label
	Kind          string              // sum or gauge
	ErrMessage    string              // resulting error message from individual metric response
}

var Metrics = []Metric{
	{
		Name:          "kv_curr_connections",
		MetadataLabel: metadata.M.CouchbaseKvCollectionDataSizeBytes,
		Kind:          "gauge",
		ValueType:     "float",
	},
	{
		Name:          "kv_collection_data_size_bytes",
		MetadataLabel: metadata.M.CouchbaseKvCollectionDataSizeBytes,
		Kind:          "gauge",
		ValueType:     "float",
	},
	{
		Name:          "kv_collection_item_count",
		MetadataLabel: metadata.M.CouchbaseKvCollectionItemCount,
		Kind:          "gauge",
		ValueType:     "float",
	},
	{
		Name:          "kv_collection_mem_used_bytes",
		MetadataLabel: metadata.M.CouchbaseKvCollectionMemUsedBytes,
		Kind:          "gauge",
		ValueType:     "float",
	},
	{
		Name:          "kv_collection_ops",
		MetadataLabel: metadata.M.CouchbaseKvCollectionOps,
		Kind:          "gauge",
		ValueType:     "float",
	},
	{
		Name:          "kv_curr_connections",
		MetadataLabel: metadata.M.CouchbaseKvCurrConnections,
		Kind:          "gauge",
		ValueType:     "float",
	},
	{
		Name:          "kv_curr_items",
		MetadataLabel: metadata.M.CouchbaseKvCurrItems,
		Kind:          "gauge",
		ValueType:     "float",
	},
	{
		Name:          "kv_curr_items_tot",
		MetadataLabel: metadata.M.CouchbaseKvCurrItemsTot,
		Kind:          "sum",
		ValueType:     "float",
	},
	// // {
	// // 	Name: "kv_item_pager_seconds_bucket",
	// // MetadataLabel: "kv_item_pager_seconds_bucket",
	// Kind: "gauge",
	// ValueType: "float",
	// // },
	// // {
	// // 	Name: "kv_item_pager_seconds_count",
	// // MetadataLabel: "kv_item_pager_seconds_count",
	// Kind: "gauge",
	// ValueType: "float",
	// // },
	// // {
	// // 	Name: "kv_item_pager_seconds_sum",
	// // MetadataLabel: "kv_item_pager_seconds_sum",
	// Kind: "gauge",
	// ValueType: "float",
	// // },
	{
		Name:          "kv_ops",
		MetadataLabel: metadata.M.CouchbaseKvOps,
		Kind:          "gauge",
		ValueType:     "float",
	},
	{
		Name:          "kv_ops_failed",
		MetadataLabel: metadata.M.CouchbaseKvOpsFailed,
		Kind:          "gauge",
		ValueType:     "float",
	},
	{
		Name:          "kv_read_bytes",
		MetadataLabel: metadata.M.CouchbaseKvReadBytes,
		Kind:          "gauge",
		ValueType:     "float",
	},
	{
		Name:          "kv_system_connections",
		MetadataLabel: metadata.M.CouchbaseKvSystemConnections,
		Kind:          "gauge",
		ValueType:     "float",
	},
	{
		Name:          "kv_time_seconds",
		MetadataLabel: metadata.M.CouchbaseKvTimeSeconds,
		Kind:          "gauge",
		ValueType:     "float",
	},
	{
		Name:          "kv_total_connections",
		MetadataLabel: metadata.M.CouchbaseKvTotalConnections,
		Kind:          "gauge",
		ValueType:     "float",
	},
	{
		Name:          "kv_total_memory_used_bytes",
		MetadataLabel: metadata.M.CouchbaseKvTotalMemoryUsedBytes,
		Kind:          "gauge",
		ValueType:     "float",
	},
	{
		Name:          "kv_total_resp_errors",
		MetadataLabel: metadata.M.CouchbaseKvTotalRespErrors,
		Kind:          "gauge",
		ValueType:     "float",
	},
	{
		Name:          "kv_uptime_seconds",
		MetadataLabel: metadata.M.CouchbaseKvUptimeSeconds,
		Kind:          "sum",
		ValueType:     "float",
	},
	{
		Name:          "kv_written_bytes",
		MetadataLabel: metadata.M.CouchbaseKvWrittenBytes,
		Kind:          "gauge",
		ValueType:     "float",
	},
}
