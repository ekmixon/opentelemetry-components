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
const Type config.Type = "mongodbreceiver"

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
	MongodbCacheHits          MetricIntf
	MongodbCacheMisses        MetricIntf
	MongodbCollections        MetricIntf
	MongodbConnections        MetricIntf
	MongodbDataSize           MetricIntf
	MongodbExtents            MetricIntf
	MongodbGlobalLockHoldTime MetricIntf
	MongodbIndexSize          MetricIntf
	MongodbIndexes            MetricIntf
	MongodbMemoryUsage        MetricIntf
	MongodbObjects            MetricIntf
	MongodbOperationCount     MetricIntf
	MongodbStorageSize        MetricIntf
}

// Names returns a list of all the metric name strings.
func (m *metricStruct) Names() []string {
	return []string{
		"mongodb.cache_hits",
		"mongodb.cache_misses",
		"mongodb.collections",
		"mongodb.connections",
		"mongodb.data_size",
		"mongodb.extents",
		"mongodb.global_lock_hold_time",
		"mongodb.index_size",
		"mongodb.indexes",
		"mongodb.memory_usage",
		"mongodb.objects",
		"mongodb.operation_count",
		"mongodb.storage_size",
	}
}

var metricsByName = map[string]MetricIntf{
	"mongodb.cache_hits":            Metrics.MongodbCacheHits,
	"mongodb.cache_misses":          Metrics.MongodbCacheMisses,
	"mongodb.collections":           Metrics.MongodbCollections,
	"mongodb.connections":           Metrics.MongodbConnections,
	"mongodb.data_size":             Metrics.MongodbDataSize,
	"mongodb.extents":               Metrics.MongodbExtents,
	"mongodb.global_lock_hold_time": Metrics.MongodbGlobalLockHoldTime,
	"mongodb.index_size":            Metrics.MongodbIndexSize,
	"mongodb.indexes":               Metrics.MongodbIndexes,
	"mongodb.memory_usage":          Metrics.MongodbMemoryUsage,
	"mongodb.objects":               Metrics.MongodbObjects,
	"mongodb.operation_count":       Metrics.MongodbOperationCount,
	"mongodb.storage_size":          Metrics.MongodbStorageSize,
}

func (m *metricStruct) ByName(n string) MetricIntf {
	return metricsByName[n]
}

func (m *metricStruct) FactoriesByName() map[string]func(pdata.Metric) {
	return map[string]func(pdata.Metric){
		Metrics.MongodbCacheHits.Name():          Metrics.MongodbCacheHits.Init,
		Metrics.MongodbCacheMisses.Name():        Metrics.MongodbCacheMisses.Init,
		Metrics.MongodbCollections.Name():        Metrics.MongodbCollections.Init,
		Metrics.MongodbConnections.Name():        Metrics.MongodbConnections.Init,
		Metrics.MongodbDataSize.Name():           Metrics.MongodbDataSize.Init,
		Metrics.MongodbExtents.Name():            Metrics.MongodbExtents.Init,
		Metrics.MongodbGlobalLockHoldTime.Name(): Metrics.MongodbGlobalLockHoldTime.Init,
		Metrics.MongodbIndexSize.Name():          Metrics.MongodbIndexSize.Init,
		Metrics.MongodbIndexes.Name():            Metrics.MongodbIndexes.Init,
		Metrics.MongodbMemoryUsage.Name():        Metrics.MongodbMemoryUsage.Init,
		Metrics.MongodbObjects.Name():            Metrics.MongodbObjects.Init,
		Metrics.MongodbOperationCount.Name():     Metrics.MongodbOperationCount.Init,
		Metrics.MongodbStorageSize.Name():        Metrics.MongodbStorageSize.Init,
	}
}

// Metrics contains a set of methods for each metric that help with
// manipulating those metrics.
var Metrics = &metricStruct{
	&metricImpl{
		"mongodb.cache_hits",
		func(metric pdata.Metric) {
			metric.SetName("mongodb.cache_hits")
			metric.SetDescription("The number of cache hits")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"mongodb.cache_misses",
		func(metric pdata.Metric) {
			metric.SetName("mongodb.cache_misses")
			metric.SetDescription("The number of cache misses")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"mongodb.collections",
		func(metric pdata.Metric) {
			metric.SetName("mongodb.collections")
			metric.SetDescription("The number of collections")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntGauge)
		},
	},
	&metricImpl{
		"mongodb.connections",
		func(metric pdata.Metric) {
			metric.SetName("mongodb.connections")
			metric.SetDescription("The number of active server connections")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntGauge)
		},
	},
	&metricImpl{
		"mongodb.data_size",
		func(metric pdata.Metric) {
			metric.SetName("mongodb.data_size")
			metric.SetDescription("The data size")
			metric.SetUnit("By")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"mongodb.extents",
		func(metric pdata.Metric) {
			metric.SetName("mongodb.extents")
			metric.SetDescription("The number of extents")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntGauge)
		},
	},
	&metricImpl{
		"mongodb.global_lock_hold_time",
		func(metric pdata.Metric) {
			metric.SetName("mongodb.global_lock_hold_time")
			metric.SetDescription("The time the global lock has been held")
			metric.SetUnit("ms")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"mongodb.index_size",
		func(metric pdata.Metric) {
			metric.SetName("mongodb.index_size")
			metric.SetDescription("The index size")
			metric.SetUnit("By")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"mongodb.indexes",
		func(metric pdata.Metric) {
			metric.SetName("mongodb.indexes")
			metric.SetDescription("The number of indexes")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntGauge)
		},
	},
	&metricImpl{
		"mongodb.memory_usage",
		func(metric pdata.Metric) {
			metric.SetName("mongodb.memory_usage")
			metric.SetDescription("The amount of memory used")
			metric.SetUnit("By")
			metric.SetDataType(pdata.MetricDataTypeIntGauge)
		},
	},
	&metricImpl{
		"mongodb.objects",
		func(metric pdata.Metric) {
			metric.SetName("mongodb.objects")
			metric.SetDescription("The number of objects")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntGauge)
		},
	},
	&metricImpl{
		"mongodb.operation_count",
		func(metric pdata.Metric) {
			metric.SetName("mongodb.operation_count")
			metric.SetDescription("The number of operations executed")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeIntSum)
			metric.IntSum().SetIsMonotonic(true)
			metric.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"mongodb.storage_size",
		func(metric pdata.Metric) {
			metric.SetName("mongodb.storage_size")
			metric.SetDescription("The storage size")
			metric.SetUnit("By")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
}

// M contains a set of methods for each metric that help with
// manipulating those metrics. M is an alias for Metrics
var M = Metrics

// Labels contains the possible metric labels that can be used.
var Labels = struct {
	// DatabaseName (The name of a database)
	DatabaseName string
	// MemoryType (The type of memory used)
	MemoryType string
	// Operation (The mongoDB operation being counted)
	Operation string
}{
	"database_name",
	"memory_type",
	"operation",
}

// L contains the possible metric labels that can be used. L is an alias for
// Labels.
var L = Labels
