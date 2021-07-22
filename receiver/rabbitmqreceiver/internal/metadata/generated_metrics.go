package metadata

import (
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/model/pdata"
)

// Type is the component type name.
const Type config.Type = "rabbitmqreceiver"

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
	RabbitmqConsumers    MetricIntf
	RabbitmqDeliveryRate MetricIntf
	RabbitmqNumMessages  MetricIntf
	RabbitmqPublishRate  MetricIntf
}

// Names returns a list of all the metric name strings.
func (m *metricStruct) Names() []string {
	return []string{
		"rabbitmq.consumers",
		"rabbitmq.delivery_rate",
		"rabbitmq.num_messages",
		"rabbitmq.publish_rate",
	}
}

var metricsByName = map[string]MetricIntf{
	"rabbitmq.consumers":     Metrics.RabbitmqConsumers,
	"rabbitmq.delivery_rate": Metrics.RabbitmqDeliveryRate,
	"rabbitmq.num_messages":  Metrics.RabbitmqNumMessages,
	"rabbitmq.publish_rate":  Metrics.RabbitmqPublishRate,
}

func (m *metricStruct) ByName(n string) MetricIntf {
	return metricsByName[n]
}

func (m *metricStruct) FactoriesByName() map[string]func(pdata.Metric) {
	return map[string]func(pdata.Metric){
		Metrics.RabbitmqConsumers.Name():    Metrics.RabbitmqConsumers.Init,
		Metrics.RabbitmqDeliveryRate.Name(): Metrics.RabbitmqDeliveryRate.Init,
		Metrics.RabbitmqNumMessages.Name():  Metrics.RabbitmqNumMessages.Init,
		Metrics.RabbitmqPublishRate.Name():  Metrics.RabbitmqPublishRate.Init,
	}
}

// Metrics contains a set of methods for each metric that help with
// manipulating those metrics.
var Metrics = &metricStruct{
	&metricImpl{
		"rabbitmq.consumers",
		func(metric pdata.Metric) {
			metric.SetName("rabbitmq.consumers")
			metric.SetDescription("The number of consumers reading from the specified queue.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"rabbitmq.delivery_rate",
		func(metric pdata.Metric) {
			metric.SetName("rabbitmq.delivery_rate")
			metric.SetDescription("The rate (per second) at which messages are being delivered.")
			metric.SetUnit("1/s")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"rabbitmq.num_messages",
		func(metric pdata.Metric) {
			metric.SetName("rabbitmq.num_messages")
			metric.SetDescription("The number of messages in a queue.")
			metric.SetUnit("1")
			metric.SetDataType(pdata.MetricDataTypeGauge)
		},
	},
	&metricImpl{
		"rabbitmq.publish_rate",
		func(metric pdata.Metric) {
			metric.SetName("rabbitmq.publish_rate")
			metric.SetDescription("The rate (per second) at which messages are being published.")
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
	// Queue (The rabbit queue name.)
	Queue string
	// State (The message state.)
	State string
}{
	"queue",
	"state",
}

// L contains the possible metric labels that can be used. L is an alias for
// Labels.
var L = Labels
