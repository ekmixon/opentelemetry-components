// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	CouchbaseKvCollectionDataSizeBytes MetricIntf
	CouchbaseKvCollectionItemCount     MetricIntf
	CouchbaseKvCollectionMemUsedBytes  MetricIntf
	CouchbaseKvCollectionOps           MetricIntf
	CouchbaseKvCurrConnections         MetricIntf
	CouchbaseKvCurrItems               MetricIntf
	CouchbaseKvCurrItemsTot            MetricIntf
	CouchbaseKvOps                     MetricIntf
	CouchbaseKvOpsFailed               MetricIntf
	CouchbaseKvReadBytes               MetricIntf
	CouchbaseKvSystemConnections       MetricIntf
	CouchbaseKvTimeSeconds             MetricIntf
	CouchbaseKvTotalConnections        MetricIntf
	CouchbaseKvTotalMemoryUsedBytes    MetricIntf
	CouchbaseKvTotalRespErrors         MetricIntf
	CouchbaseKvUptimeSeconds           MetricIntf
	CouchbaseKvWrittenBytes            MetricIntf
}

// Names returns a list of all the metric name strings.
func (m *metricStruct) Names() []string {
	return []string{
		"couchbase.kv_collection_data_size_bytes",
		"couchbase.kv_collection_item_count",
		"couchbase.kv_collection_mem_used_bytes",
		"couchbase.kv_collection_ops",
		"couchbase.kv_curr_connections",
		"couchbase.kv_curr_items",
		"couchbase.kv_curr_items_tot",
		"couchbase.kv_ops",
		"couchbase.kv_ops_failed",
		"couchbase.kv_read_bytes",
		"couchbase.kv_system_connections",
		"couchbase.kv_time_seconds",
		"couchbase.kv_total_connections",
		"couchbase.kv_total_memory_used_bytes",
		"couchbase.kv_total_resp_errors",
		"couchbase.kv_uptime_seconds",
		"couchbase.kv_written_bytes",
	}
}

var metricsByName = map[string]MetricIntf{
	"couchbase.kv_collection_data_size_bytes": Metrics.CouchbaseKvCollectionDataSizeBytes,
	"couchbase.kv_collection_item_count":      Metrics.CouchbaseKvCollectionItemCount,
	"couchbase.kv_collection_mem_used_bytes":  Metrics.CouchbaseKvCollectionMemUsedBytes,
	"couchbase.kv_collection_ops":             Metrics.CouchbaseKvCollectionOps,
	"couchbase.kv_curr_connections":           Metrics.CouchbaseKvCurrConnections,
	"couchbase.kv_curr_items":                 Metrics.CouchbaseKvCurrItems,
	"couchbase.kv_curr_items_tot":             Metrics.CouchbaseKvCurrItemsTot,
	"couchbase.kv_ops":                        Metrics.CouchbaseKvOps,
	"couchbase.kv_ops_failed":                 Metrics.CouchbaseKvOpsFailed,
	"couchbase.kv_read_bytes":                 Metrics.CouchbaseKvReadBytes,
	"couchbase.kv_system_connections":         Metrics.CouchbaseKvSystemConnections,
	"couchbase.kv_time_seconds":               Metrics.CouchbaseKvTimeSeconds,
	"couchbase.kv_total_connections":          Metrics.CouchbaseKvTotalConnections,
	"couchbase.kv_total_memory_used_bytes":    Metrics.CouchbaseKvTotalMemoryUsedBytes,
	"couchbase.kv_total_resp_errors":          Metrics.CouchbaseKvTotalRespErrors,
	"couchbase.kv_uptime_seconds":             Metrics.CouchbaseKvUptimeSeconds,
	"couchbase.kv_written_bytes":              Metrics.CouchbaseKvWrittenBytes,
}

func (m *metricStruct) ByName(n string) MetricIntf {
	return metricsByName[n]
}

func (m *metricStruct) FactoriesByName() map[string]func(pdata.Metric) {
	return map[string]func(pdata.Metric){
		Metrics.CouchbaseKvCollectionDataSizeBytes.Name(): Metrics.CouchbaseKvCollectionDataSizeBytes.Init,
		Metrics.CouchbaseKvCollectionItemCount.Name():     Metrics.CouchbaseKvCollectionItemCount.Init,
		Metrics.CouchbaseKvCollectionMemUsedBytes.Name():  Metrics.CouchbaseKvCollectionMemUsedBytes.Init,
		Metrics.CouchbaseKvCollectionOps.Name():           Metrics.CouchbaseKvCollectionOps.Init,
		Metrics.CouchbaseKvCurrConnections.Name():         Metrics.CouchbaseKvCurrConnections.Init,
		Metrics.CouchbaseKvCurrItems.Name():               Metrics.CouchbaseKvCurrItems.Init,
		Metrics.CouchbaseKvCurrItemsTot.Name():            Metrics.CouchbaseKvCurrItemsTot.Init,
		Metrics.CouchbaseKvOps.Name():                     Metrics.CouchbaseKvOps.Init,
		Metrics.CouchbaseKvOpsFailed.Name():               Metrics.CouchbaseKvOpsFailed.Init,
		Metrics.CouchbaseKvReadBytes.Name():               Metrics.CouchbaseKvReadBytes.Init,
		Metrics.CouchbaseKvSystemConnections.Name():       Metrics.CouchbaseKvSystemConnections.Init,
		Metrics.CouchbaseKvTimeSeconds.Name():             Metrics.CouchbaseKvTimeSeconds.Init,
		Metrics.CouchbaseKvTotalConnections.Name():        Metrics.CouchbaseKvTotalConnections.Init,
		Metrics.CouchbaseKvTotalMemoryUsedBytes.Name():    Metrics.CouchbaseKvTotalMemoryUsedBytes.Init,
		Metrics.CouchbaseKvTotalRespErrors.Name():         Metrics.CouchbaseKvTotalRespErrors.Init,
		Metrics.CouchbaseKvUptimeSeconds.Name():           Metrics.CouchbaseKvUptimeSeconds.Init,
		Metrics.CouchbaseKvWrittenBytes.Name():            Metrics.CouchbaseKvWrittenBytes.Init,
	}
}

// Metrics contains a set of methods for each metric that help with
// manipulating those metrics.
var Metrics = &metricStruct{
	&metricImpl{
		"couchbase.kv_collection_data_size_bytes",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.kv_collection_data_size_bytes")
			metric.SetDescription("The kv_collection_data_size_bytes")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.kv_collection_item_count",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.kv_collection_item_count")
			metric.SetDescription("The kv_collection_item_count")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.kv_collection_mem_used_bytes",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.kv_collection_mem_used_bytes")
			metric.SetDescription("The kv_collection_mem_used_bytes")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.kv_collection_ops",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.kv_collection_ops")
			metric.SetDescription("The kv_collection_ops")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.kv_curr_connections",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.kv_curr_connections")
			metric.SetDescription("The kv_curr_connections")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.kv_curr_items",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.kv_curr_items")
			metric.SetDescription("The kv_curr_items")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.kv_curr_items_tot",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.kv_curr_items_tot")
			metric.SetDescription("The kv_curr_items_tot")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeSum)
			metric.Sum().SetIsMonotonic(true)
			metric.Sum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"couchbase.kv_ops",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.kv_ops")
			metric.SetDescription("The kv_ops")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.kv_ops_failed",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.kv_ops_failed")
			metric.SetDescription("The kv_ops_failed")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.kv_read_bytes",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.kv_read_bytes")
			metric.SetDescription("The kv_read_bytes")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.kv_system_connections",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.kv_system_connections")
			metric.SetDescription("The kv_system_connections")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.kv_time_seconds",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.kv_time_seconds")
			metric.SetDescription("The kv_time_seconds")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.kv_total_connections",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.kv_total_connections")
			metric.SetDescription("The kv_total_connections")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.kv_total_memory_used_bytes",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.kv_total_memory_used_bytes")
			metric.SetDescription("The kv_total_memory_used_bytes")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.kv_total_resp_errors",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.kv_total_resp_errors")
			metric.SetDescription("The kv_total_resp_errors")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchbase.kv_uptime_seconds",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.kv_uptime_seconds")
			metric.SetDescription("The kv_uptime_seconds")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeSum)
			metric.Sum().SetIsMonotonic(true)
			metric.Sum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"couchbase.kv_written_bytes",
		func(metric pdata.Metric) {
			metric.SetName("couchbase.kv_written_bytes")
			metric.SetDescription("The kv_written_bytes")
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
}{}

// L contains the possible metric labels that can be used. L is an alias for
// Labels.
var L = Labels
