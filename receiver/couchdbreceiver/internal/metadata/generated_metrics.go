// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/model/pdata"
)

// Type is the component type name.
const Type config.Type = "couchdbreceiver"

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
	CouchdbHttpdBulkRequests       MetricIntf
	CouchdbHttpdRequestMethods     MetricIntf
	CouchdbHttpdResponseCodes      MetricIntf
	CouchdbHttpdTemporaryViewReads MetricIntf
	CouchdbOpenDatabases           MetricIntf
	CouchdbOpenFiles               MetricIntf
	CouchdbReads                   MetricIntf
	CouchdbRequestTime             MetricIntf
	CouchdbViewReads               MetricIntf
	CouchdbWrites                  MetricIntf
}

// Names returns a list of all the metric name strings.
func (m *metricStruct) Names() []string {
	return []string{
		"couchdb.httpd_bulk_requests",
		"couchdb.httpd_request_methods",
		"couchdb.httpd_response_codes",
		"couchdb.httpd_temporary_view_reads",
		"couchdb.open_databases",
		"couchdb.open_files",
		"couchdb.reads",
		"couchdb.request_time",
		"couchdb.view_reads",
		"couchdb.writes",
	}
}

var metricsByName = map[string]MetricIntf{
	"couchdb.httpd_bulk_requests":        Metrics.CouchdbHttpdBulkRequests,
	"couchdb.httpd_request_methods":      Metrics.CouchdbHttpdRequestMethods,
	"couchdb.httpd_response_codes":       Metrics.CouchdbHttpdResponseCodes,
	"couchdb.httpd_temporary_view_reads": Metrics.CouchdbHttpdTemporaryViewReads,
	"couchdb.open_databases":             Metrics.CouchdbOpenDatabases,
	"couchdb.open_files":                 Metrics.CouchdbOpenFiles,
	"couchdb.reads":                      Metrics.CouchdbReads,
	"couchdb.request_time":               Metrics.CouchdbRequestTime,
	"couchdb.view_reads":                 Metrics.CouchdbViewReads,
	"couchdb.writes":                     Metrics.CouchdbWrites,
}

func (m *metricStruct) ByName(n string) MetricIntf {
	return metricsByName[n]
}

func (m *metricStruct) FactoriesByName() map[string]func(pdata.Metric) {
	return map[string]func(pdata.Metric){
		Metrics.CouchdbHttpdBulkRequests.Name():       Metrics.CouchdbHttpdBulkRequests.Init,
		Metrics.CouchdbHttpdRequestMethods.Name():     Metrics.CouchdbHttpdRequestMethods.Init,
		Metrics.CouchdbHttpdResponseCodes.Name():      Metrics.CouchdbHttpdResponseCodes.Init,
		Metrics.CouchdbHttpdTemporaryViewReads.Name(): Metrics.CouchdbHttpdTemporaryViewReads.Init,
		Metrics.CouchdbOpenDatabases.Name():           Metrics.CouchdbOpenDatabases.Init,
		Metrics.CouchdbOpenFiles.Name():               Metrics.CouchdbOpenFiles.Init,
		Metrics.CouchdbReads.Name():                   Metrics.CouchdbReads.Init,
		Metrics.CouchdbRequestTime.Name():             Metrics.CouchdbRequestTime.Init,
		Metrics.CouchdbViewReads.Name():               Metrics.CouchdbViewReads.Init,
		Metrics.CouchdbWrites.Name():                  Metrics.CouchdbWrites.Init,
	}
}

// Metrics contains a set of methods for each metric that help with
// manipulating those metrics.
var Metrics = &metricStruct{
	&metricImpl{
		"couchdb.httpd_bulk_requests",
		func(metric pdata.Metric) {
			metric.SetName("couchdb.httpd_bulk_requests")
			metric.SetDescription("The bulk request count.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeSum)
			metric.Sum().SetIsMonotonic(true)
			metric.Sum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"couchdb.httpd_request_methods",
		func(metric pdata.Metric) {
			metric.SetName("couchdb.httpd_request_methods")
			metric.SetDescription("The HTTP request method count.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeSum)
			metric.Sum().SetIsMonotonic(true)
			metric.Sum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"couchdb.httpd_response_codes",
		func(metric pdata.Metric) {
			metric.SetName("couchdb.httpd_response_codes")
			metric.SetDescription("The HTTP request method count.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeSum)
			metric.Sum().SetIsMonotonic(true)
			metric.Sum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"couchdb.httpd_temporary_view_reads",
		func(metric pdata.Metric) {
			metric.SetName("couchdb.httpd_temporary_view_reads")
			metric.SetDescription("The temporary view reads count.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeSum)
			metric.Sum().SetIsMonotonic(true)
			metric.Sum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"couchdb.open_databases",
		func(metric pdata.Metric) {
			metric.SetName("couchdb.open_databases")
			metric.SetDescription("The number of open databases.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchdb.open_files",
		func(metric pdata.Metric) {
			metric.SetName("couchdb.open_files")
			metric.SetDescription("The number of open files.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchdb.reads",
		func(metric pdata.Metric) {
			metric.SetName("couchdb.reads")
			metric.SetDescription("The database read count.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeSum)
			metric.Sum().SetIsMonotonic(true)
			metric.Sum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"couchdb.request_time",
		func(metric pdata.Metric) {
			metric.SetName("couchdb.request_time")
			metric.SetDescription("The average request time.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"couchdb.view_reads",
		func(metric pdata.Metric) {
			metric.SetName("couchdb.view_reads")
			metric.SetDescription("The view reads count.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeSum)
			metric.Sum().SetIsMonotonic(true)
			metric.Sum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		},
	},
	&metricImpl{
		"couchdb.writes",
		func(metric pdata.Metric) {
			metric.SetName("couchdb.writes")
			metric.SetDescription("The database write count.")
			metric.SetUnit("1")
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
	// HTTPMethod (An HTTP request method.)
	HTTPMethod string
	// NodeName (The name of the node.)
	NodeName string
	// ResponseCode (An HTTP status code.)
	ResponseCode string
}{
	"http_method",
	"node_name",
	"response_code",
}

// L contains the possible metric labels that can be used. L is an alias for
// Labels.
var L = Labels

// LabelHTTPMethod are the possible values that the label "http_method" can have.
var LabelHTTPMethod = struct {
	COPY    string
	DELETE  string
	GET     string
	HEAD    string
	OPTIONS string
	POST    string
	PUT     string
}{
	"COPY",
	"DELETE",
	"GET",
	"HEAD",
	"OPTIONS",
	"POST",
	"PUT",
}

// LabelResponseCode are the possible values that the label "response_code" can have.
var LabelResponseCode = struct {
	HTTP200 string
	HTTP201 string
	HTTP202 string
	HTTP204 string
	HTTP206 string
	HTTP301 string
	HTTP302 string
	HTTP304 string
	HTTP400 string
	HTTP401 string
	HTTP403 string
	HTTP404 string
	HTTP405 string
	HTTP406 string
	HTTP409 string
	HTTP412 string
	HTTP413 string
	HTTP414 string
	HTTP415 string
	HTTP416 string
	HTTP417 string
	HTTP500 string
	HTTP501 string
	HTTP503 string
}{
	"http_200",
	"http_201",
	"http_202",
	"http_204",
	"http_206",
	"http_301",
	"http_302",
	"http_304",
	"http_400",
	"http_401",
	"http_403",
	"http_404",
	"http_405",
	"http_406",
	"http_409",
	"http_412",
	"http_413",
	"http_414",
	"http_415",
	"http_416",
	"http_417",
	"http_500",
	"http_501",
	"http_503",
}
