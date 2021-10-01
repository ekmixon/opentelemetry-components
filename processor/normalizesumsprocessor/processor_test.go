package normalizesumsprocessor

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/collector/processor/processorhelper"
	"go.uber.org/zap"
)

type testCase struct {
	name     string
	inputs   []pdata.Metrics
	expected []pdata.Metrics
}

func TestNormalizeSumsProcessor(t *testing.T) {
	testStart := time.Now().Unix()
	tests := []testCase{
		{
			name:     "no-transform-case",
			inputs:   generateNoTransformMetrics(testStart),
			expected: generateNoTransformMetrics(testStart),
		},
		{
			name:     "removed-metric-case",
			inputs:   generateRemoveInput(testStart),
			expected: generateRemoveOutput(testStart),
		},
		{
			name:     "transform-all-happy-case",
			inputs:   generateLabelledInput(testStart),
			expected: generateLabelledOutput(testStart),
		},
		{
			name:     "transform-all-label-separated-case",
			inputs:   generateSeparatedLabelledInput(testStart),
			expected: generateSeparatedLabelledOutput(testStart),
		},
		{
			name:     "more-complex-case",
			inputs:   generateComplexInput(testStart),
			expected: generateComplexOutput(testStart),
		},
		{
			name:     "multiple-resource-case",
			inputs:   generateMultipleResourceInput(testStart),
			expected: generateMultipleResourceOutput(testStart),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nsp := newNormalizeSumsProcessor(zap.NewExample())

			tmn := &consumertest.MetricsSink{}
			id := config.NewID(typeStr)
			settings := config.NewProcessorSettings(id)
			rmp, err := processorhelper.NewMetricsProcessor(
				&Config{
					ProcessorSettings: &settings,
				},
				tmn,
				nsp.ProcessMetrics,
				processorhelper.WithCapabilities(processorCapabilities))
			require.NoError(t, err)

			require.True(t, rmp.Capabilities().MutatesData)

			require.NoError(t, rmp.Start(context.Background(), componenttest.NewNopHost()))
			defer func() { require.NoError(t, rmp.Shutdown(context.Background())) }()

			for _, input := range tt.inputs {
				err = rmp.ConsumeMetrics(context.Background(), input)
				require.NoError(t, err)
			}

			requireEqual(t, tt.expected, tmn.AllMetrics())
		})
	}
}

func generateNoTransformMetrics(startTime int64) []pdata.Metrics {
	input := pdata.NewMetrics()

	rmb := newResourceMetricsBuilder()
	b := rmb.addResourceMetrics(nil)

	mb1 := b.addMetric("m1", pdata.MetricDataTypeSum, true)
	mb1.addDoubleDataPoint(1, map[string]pdata.AttributeValue{}, startTime+1000, startTime)
	mb1.addDoubleDataPoint(5, map[string]pdata.AttributeValue{}, startTime+2000, startTime)
	mb1.addDoubleDataPoint(2, map[string]pdata.AttributeValue{}, startTime+3000, startTime+2000)

	mb2 := b.addMetric("m2", pdata.MetricDataTypeSum, true)
	mb2.addDoubleDataPoint(3, map[string]pdata.AttributeValue{}, startTime+6000, startTime)
	mb2.addDoubleDataPoint(4, map[string]pdata.AttributeValue{}, startTime+7000, startTime)

	mb3 := b.addMetric("m3", pdata.MetricDataTypeGauge, false)
	mb3.addIntDataPoint(5, map[string]pdata.AttributeValue{}, startTime, 0)
	mb3.addIntDataPoint(4, map[string]pdata.AttributeValue{}, startTime+1000, 0)

	mb4 := b.addMetric("m4", pdata.MetricDataTypeGauge, false)
	mb4.addDoubleDataPoint(50000.2, map[string]pdata.AttributeValue{}, startTime, 0)
	mb4.addDoubleDataPoint(11, map[string]pdata.AttributeValue{}, startTime+1000, 0)

	rmb.Build().CopyTo(input.ResourceMetrics())
	return []pdata.Metrics{input}
}

func generateMultipleResourceInput(startTime int64) []pdata.Metrics {
	input := pdata.NewMetrics()

	rmb := newResourceMetricsBuilder()
	b := rmb.addResourceMetrics(map[string]pdata.AttributeValue{
		"label1": pdata.NewAttributeValueString("value1"),
	})

	mb1 := b.addMetric("m1", pdata.MetricDataTypeSum, true)
	mb1.addDoubleDataPoint(1, map[string]pdata.AttributeValue{}, startTime, 0)
	mb1.addDoubleDataPoint(2, map[string]pdata.AttributeValue{}, startTime+1000, 0)

	b2 := rmb.addResourceMetrics(map[string]pdata.AttributeValue{
		"label1": pdata.NewAttributeValueString("value2"),
	})

	mb2 := b2.addMetric("m1", pdata.MetricDataTypeSum, true)
	mb2.addDoubleDataPoint(5, map[string]pdata.AttributeValue{}, startTime+2000, 0)
	mb2.addDoubleDataPoint(10, map[string]pdata.AttributeValue{}, startTime+3000, 0)

	rmb.Build().CopyTo(input.ResourceMetrics())
	return []pdata.Metrics{input}
}

func generateMultipleResourceOutput(startTime int64) []pdata.Metrics {
	output := pdata.NewMetrics()

	rmb := newResourceMetricsBuilder()
	b := rmb.addResourceMetrics(map[string]pdata.AttributeValue{
		"label1": pdata.NewAttributeValueString("value1"),
	})

	mb1 := b.addMetric("m1", pdata.MetricDataTypeSum, true)
	//mb1.addDoubleDataPoint(1, map[string]pdata.AttributeValue{}, startTime, 0)
	mb1.addDoubleDataPoint(1, map[string]pdata.AttributeValue{}, startTime+1000, startTime)

	b2 := rmb.addResourceMetrics(map[string]pdata.AttributeValue{
		"label1": pdata.NewAttributeValueString("value2"),
	})

	mb2 := b2.addMetric("m1", pdata.MetricDataTypeSum, true)
	//mb2.addDoubleDataPoint(5, map[string]pdata.AttributeValue{}, startTime+2000, 0)
	mb2.addDoubleDataPoint(5, map[string]pdata.AttributeValue{}, startTime+3000, startTime+2000)

	rmb.Build().CopyTo(output.ResourceMetrics())
	return []pdata.Metrics{output}
}

func generateLabelledInput(startTime int64) []pdata.Metrics {
	input := pdata.NewMetrics()

	rmb := newResourceMetricsBuilder()
	b := rmb.addResourceMetrics(nil)

	mb1 := b.addMetric("m1", pdata.MetricDataTypeSum, true)
	mb1.addDoubleDataPoint(0, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val1"),
	}, startTime, 0)
	mb1.addDoubleDataPoint(3, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val2"),
	}, startTime, 0)
	mb1.addDoubleDataPoint(12, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val1"),
	}, startTime+1000, 0)
	mb1.addDoubleDataPoint(5, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val2"),
	}, startTime+1000, 0)
	mb1.addDoubleDataPoint(15, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val1"),
	}, startTime+2000, 0)
	mb1.addDoubleDataPoint(1, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val2"),
	}, startTime+2000, 0)
	mb1.addDoubleDataPoint(22, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val1"),
	}, startTime+3000, 0)
	mb1.addDoubleDataPoint(11, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val2"),
	}, startTime+3000, 0)

	rmb.Build().CopyTo(input.ResourceMetrics())
	return []pdata.Metrics{input}
}

func generateLabelledOutput(startTime int64) []pdata.Metrics {
	output := pdata.NewMetrics()

	rmb := newResourceMetricsBuilder()
	b := rmb.addResourceMetrics(nil)

	mb1 := b.addMetric("m1", pdata.MetricDataTypeSum, true)
	// mb1.addDoubleDataPoint(1, map[string]pdata.AttributeValue{
	// "label": pdata.NewAttributeValueString("val1"),
	// }, startTime, 0)
	// mb1.addDoubleDataPoint(1, map[string]pdata.AttributeValue{
	// "label": pdata.NewAttributeValueString("val2"),
	// }, startTime, 0)
	mb1.addDoubleDataPoint(12, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val1"),
	}, startTime+1000, startTime)
	mb1.addDoubleDataPoint(2, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val2"),
	}, startTime+1000, startTime)
	mb1.addDoubleDataPoint(15, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val1"),
	}, startTime+2000, startTime)
	//mb1.addDoubleDataPoint(1, map[string]pdata.AttributeValue{
	// "label": pdata.NewAttributeValueString("val2"),
	// }, startTime+2000, 1)
	mb1.addDoubleDataPoint(22, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val1"),
	}, startTime+3000, startTime)
	mb1.addDoubleDataPoint(10, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val2"),
	}, startTime+3000, startTime+2000)

	rmb.Build().CopyTo(output.ResourceMetrics())
	return []pdata.Metrics{output}
}

func generateSeparatedLabelledInput(startTime int64) []pdata.Metrics {
	input := pdata.NewMetrics()

	rmb := newResourceMetricsBuilder()
	b := rmb.addResourceMetrics(nil)

	mb1 := b.addMetric("m1", pdata.MetricDataTypeSum, true)
	mb1.addDoubleDataPoint(0, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val1"),
	}, startTime, 0)
	mb1.addDoubleDataPoint(12, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val1"),
	}, startTime+1000, 0)
	mb1.addDoubleDataPoint(15, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val1"),
	}, startTime+2000, 0)
	mb1.addDoubleDataPoint(22, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val1"),
	}, startTime+3000, 0)

	mb2 := b.addMetric("m1", pdata.MetricDataTypeSum, true)
	mb2.addDoubleDataPoint(3, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val2"),
	}, startTime, 0)
	mb2.addDoubleDataPoint(5, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val2"),
	}, startTime+1000, 0)
	mb2.addDoubleDataPoint(1, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val2"),
	}, startTime+2000, 0)
	mb2.addDoubleDataPoint(11, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val2"),
	}, startTime+3000, 0)

	rmb.Build().CopyTo(input.ResourceMetrics())
	return []pdata.Metrics{input}
}

func generateSeparatedLabelledOutput(startTime int64) []pdata.Metrics {
	output := pdata.NewMetrics()

	rmb := newResourceMetricsBuilder()
	b := rmb.addResourceMetrics(nil)

	mb1 := b.addMetric("m1", pdata.MetricDataTypeSum, true)
	// mb1.addDoubleDataPoint(1, map[string]pdata.AttributeValue{
	// "label": pdata.NewAttributeValueString("val1"),
	// }, startTime, 0)
	mb1.addDoubleDataPoint(12, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val1"),
	}, startTime+1000, startTime)
	mb1.addDoubleDataPoint(15, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val1"),
	}, startTime+2000, startTime)
	mb1.addDoubleDataPoint(22, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val1"),
	}, startTime+3000, startTime)

	mb2 := b.addMetric("m1", pdata.MetricDataTypeSum, true)
	// mb2.addDoubleDataPoint(1, map[string]pdata.AttributeValue{
	// "label": pdata.NewAttributeValueString("val2"),
	// }, startTime, 0)
	mb2.addDoubleDataPoint(2, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val2"),
	}, startTime+1000, startTime)
	//mb2.addDoubleDataPoint(1, map[string]pdata.AttributeValue{
	// "label": pdata.NewAttributeValueString("val2"),
	// }, startTime+2000, 1)
	mb2.addDoubleDataPoint(10, map[string]pdata.AttributeValue{
		"label": pdata.NewAttributeValueString("val2"),
	}, startTime+3000, startTime+2000)

	rmb.Build().CopyTo(output.ResourceMetrics())
	return []pdata.Metrics{output}
}

func generateRemoveInput(startTime int64) []pdata.Metrics {
	input := pdata.NewMetrics()

	rmb := newResourceMetricsBuilder()
	b := rmb.addResourceMetrics(nil)

	mb1 := b.addMetric("m1", pdata.MetricDataTypeSum, true)
	mb1.addDoubleDataPoint(1, map[string]pdata.AttributeValue{}, startTime, 0)

	mb2 := b.addMetric("m2", pdata.MetricDataTypeSum, true)
	mb2.addIntDataPoint(3, map[string]pdata.AttributeValue{}, startTime, 0)
	mb2.addIntDataPoint(4, map[string]pdata.AttributeValue{}, startTime+1000, 0)

	mb3 := b.addMetric("m3", pdata.MetricDataTypeGauge, false)
	mb3.addDoubleDataPoint(5, map[string]pdata.AttributeValue{}, startTime, 0)
	mb3.addDoubleDataPoint(6, map[string]pdata.AttributeValue{}, startTime+1000, 0)

	rmb.Build().CopyTo(input.ResourceMetrics())
	return []pdata.Metrics{input}
}

func generateRemoveOutput(startTime int64) []pdata.Metrics {
	output := pdata.NewMetrics()

	rmb := newResourceMetricsBuilder()
	b := rmb.addResourceMetrics(nil)

	mb2 := b.addMetric("m2", pdata.MetricDataTypeSum, true)
	//mb2.addIntDataPoint(3, map[string]pdata.AttributeValue{}, startTime, 0)
	mb2.addIntDataPoint(1, map[string]pdata.AttributeValue{}, startTime+1000, startTime)

	mb3 := b.addMetric("m3", pdata.MetricDataTypeGauge, false)
	mb3.addDoubleDataPoint(5, map[string]pdata.AttributeValue{}, startTime, 0)
	mb3.addDoubleDataPoint(6, map[string]pdata.AttributeValue{}, startTime+1000, 0)

	rmb.Build().CopyTo(output.ResourceMetrics())
	return []pdata.Metrics{output}
}

func generateComplexInput(startTime int64) []pdata.Metrics {
	list := []pdata.Metrics{}
	input := pdata.NewMetrics()

	rmb := newResourceMetricsBuilder()
	b := rmb.addResourceMetrics(nil)

	mb1 := b.addMetric("m1", pdata.MetricDataTypeSum, true)
	mb1.addDoubleDataPoint(1, map[string]pdata.AttributeValue{}, startTime, 0)
	mb1.addDoubleDataPoint(2, map[string]pdata.AttributeValue{}, startTime+1000, 0)
	mb1.addDoubleDataPoint(2, map[string]pdata.AttributeValue{}, startTime+2000, 0)
	mb1.addDoubleDataPoint(5, map[string]pdata.AttributeValue{}, startTime+3000, 0)
	mb1.addDoubleDataPoint(2, map[string]pdata.AttributeValue{}, startTime+4000, 0)
	mb1.addDoubleDataPoint(4, map[string]pdata.AttributeValue{}, startTime+5000, 0)

	mb2 := b.addMetric("m2", pdata.MetricDataTypeSum, true)
	mb2.addIntDataPoint(3, map[string]pdata.AttributeValue{}, startTime, 0)
	mb2.addIntDataPoint(4, map[string]pdata.AttributeValue{}, startTime+1000, 0)
	mb2.addIntDataPoint(5, map[string]pdata.AttributeValue{}, startTime+2000, 0)
	mb2.addIntDataPoint(2, map[string]pdata.AttributeValue{}, startTime, 0)
	mb2.addIntDataPoint(8, map[string]pdata.AttributeValue{}, startTime+3000, 0)
	mb2.addIntDataPoint(2, map[string]pdata.AttributeValue{}, startTime+10000, 0)
	mb2.addIntDataPoint(6, map[string]pdata.AttributeValue{}, startTime+120000, 0)

	mb3 := b.addMetric("m3", pdata.MetricDataTypeGauge, false)
	mb3.addIntDataPoint(5, map[string]pdata.AttributeValue{}, startTime, 0)
	mb3.addIntDataPoint(6, map[string]pdata.AttributeValue{}, startTime+1000, 0)

	rmb.Build().CopyTo(input.ResourceMetrics())
	list = append(list, input)

	input = pdata.NewMetrics()
	rmb = newResourceMetricsBuilder()
	b = rmb.addResourceMetrics(nil)

	mb1 = b.addMetric("m1", pdata.MetricDataTypeSum, true)
	mb1.addDoubleDataPoint(7, map[string]pdata.AttributeValue{}, startTime+6000, 0)
	mb1.addDoubleDataPoint(9, map[string]pdata.AttributeValue{}, startTime+7000, 0)

	rmb.Build().CopyTo(input.ResourceMetrics())
	list = append(list, input)

	return list
}

func generateComplexOutput(startTime int64) []pdata.Metrics {
	list := []pdata.Metrics{}
	output := pdata.NewMetrics()

	rmb := newResourceMetricsBuilder()
	b := rmb.addResourceMetrics(nil)

	mb1 := b.addMetric("m1", pdata.MetricDataTypeSum, true)
	// mb1.addDoubleDataPoint(1, map[string]pdata.AttributeValue{}, startTime, 0)
	mb1.addDoubleDataPoint(1, map[string]pdata.AttributeValue{}, startTime+1000, startTime)
	mb1.addDoubleDataPoint(1, map[string]pdata.AttributeValue{}, startTime+2000, startTime)
	mb1.addDoubleDataPoint(4, map[string]pdata.AttributeValue{}, startTime+3000, startTime)
	// mb1.addDoubleDataPoint(2, map[string]pdata.AttributeValue{}, startTime+4000, 0)
	mb1.addDoubleDataPoint(2, map[string]pdata.AttributeValue{}, startTime+5000, startTime+4000)

	mb2 := b.addMetric("m2", pdata.MetricDataTypeSum, true)
	// mb2.addIntDataPoint(3, map[string]pdata.AttributeValue{}, startTime, 0)
	mb2.addIntDataPoint(1, map[string]pdata.AttributeValue{}, startTime+1000, startTime)
	mb2.addIntDataPoint(2, map[string]pdata.AttributeValue{}, startTime+2000, startTime)
	// mb2.addIntDataPoint(2, map[string]pdata.AttributeValue{}, startTime, 0)
	mb2.addIntDataPoint(5, map[string]pdata.AttributeValue{}, startTime+3000, startTime)
	// mb2.addIntDataPoint(2, map[string]pdata.AttributeValue{}, startTime+10000, 0)
	mb2.addIntDataPoint(4, map[string]pdata.AttributeValue{}, startTime+120000, startTime+10000)

	mb3 := b.addMetric("m3", pdata.MetricDataTypeGauge, false)
	mb3.addIntDataPoint(5, map[string]pdata.AttributeValue{}, startTime, 0)
	mb3.addIntDataPoint(6, map[string]pdata.AttributeValue{}, startTime+1000, 0)

	rmb.Build().CopyTo(output.ResourceMetrics())
	list = append(list, output)

	output = pdata.NewMetrics()

	rmb = newResourceMetricsBuilder()
	b = rmb.addResourceMetrics(nil)

	mb1 = b.addMetric("m1", pdata.MetricDataTypeSum, true)
	mb1.addDoubleDataPoint(5, map[string]pdata.AttributeValue{}, startTime+6000, startTime+4000)
	mb1.addDoubleDataPoint(7, map[string]pdata.AttributeValue{}, startTime+7000, startTime+4000)

	rmb.Build().CopyTo(output.ResourceMetrics())
	list = append(list, output)

	return list
}

// builders to generate test metrics

type resourceMetricsBuilder struct {
	rms pdata.ResourceMetricsSlice
}

func newResourceMetricsBuilder() resourceMetricsBuilder {
	return resourceMetricsBuilder{rms: pdata.NewResourceMetricsSlice()}
}

func (rmsb resourceMetricsBuilder) addResourceMetrics(resourceAttributes map[string]pdata.AttributeValue) metricsBuilder {
	rm := rmsb.rms.AppendEmpty()

	if resourceAttributes != nil {
		rm.Resource().Attributes().InitFromMap(resourceAttributes)
	}

	ilm := rm.InstrumentationLibraryMetrics().AppendEmpty()

	return metricsBuilder{metrics: ilm.Metrics()}
}

func (rmsb resourceMetricsBuilder) Build() pdata.ResourceMetricsSlice {
	return rmsb.rms
}

type metricsBuilder struct {
	metrics pdata.MetricSlice
}

func (msb metricsBuilder) addMetric(name string, t pdata.MetricDataType, isMonotonic bool) metricBuilder {
	metric := msb.metrics.AppendEmpty()
	metric.SetName(name)
	metric.SetDataType(t)

	switch t {
	case pdata.MetricDataTypeSum:
		sum := metric.Sum()
		sum.SetIsMonotonic(isMonotonic)
		sum.SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
	}

	return metricBuilder{metric: metric}
}

type metricBuilder struct {
	metric pdata.Metric
}

func (mb metricBuilder) addDoubleDataPoint(value float64, attributes map[string]pdata.AttributeValue, timestamp int64, startTimestamp int64) {
	switch mb.metric.DataType() {
	case pdata.MetricDataTypeSum:
		ddp := mb.metric.Sum().DataPoints().AppendEmpty()
		ddp.Attributes().InitFromMap(attributes)
		ddp.SetDoubleVal(value)
		ddp.SetTimestamp(pdata.NewTimestampFromTime(time.Unix(timestamp, 0)))
		if startTimestamp > 0 {
			ddp.SetStartTimestamp(pdata.NewTimestampFromTime(time.Unix(startTimestamp, 0)))
		}
	case pdata.MetricDataTypeGauge:
		ddp := mb.metric.Gauge().DataPoints().AppendEmpty()
		ddp.Attributes().InitFromMap(attributes)
		ddp.SetDoubleVal(value)
		ddp.SetTimestamp(pdata.NewTimestampFromTime(time.Unix(timestamp, 0)))
		if startTimestamp > 0 {
			ddp.SetStartTimestamp(pdata.NewTimestampFromTime(time.Unix(startTimestamp, 0)))
		}
	}
}

func (mb metricBuilder) addIntDataPoint(value int64, attributes map[string]pdata.AttributeValue, timestamp int64, startTimestamp int64) {
	switch mb.metric.DataType() {
	case pdata.MetricDataTypeSum:
		ddp := mb.metric.Sum().DataPoints().AppendEmpty()
		ddp.Attributes().InitFromMap(attributes)
		ddp.SetIntVal(value)
		ddp.SetTimestamp(pdata.NewTimestampFromTime(time.Unix(timestamp, 0)))
		if startTimestamp > 0 {
			ddp.SetStartTimestamp(pdata.NewTimestampFromTime(time.Unix(startTimestamp, 0)))
		}
	case pdata.MetricDataTypeGauge:
		ddp := mb.metric.Gauge().DataPoints().AppendEmpty()
		ddp.Attributes().InitFromMap(attributes)
		ddp.SetIntVal(value)
		ddp.SetTimestamp(pdata.NewTimestampFromTime(time.Unix(timestamp, 0)))
		if startTimestamp > 0 {
			ddp.SetStartTimestamp(pdata.NewTimestampFromTime(time.Unix(startTimestamp, 0)))
		}
	}
}

// requireEqual is required because Attribute & Label Maps are not sorted by default
// and we don't provide any guarantees on the order of transformed metrics
func requireEqual(t *testing.T, expected, actual []pdata.Metrics) {
	require.Equal(t, len(expected), len(actual))

	for q := 0; q < len(actual); q++ {
		rmsAct := actual[q].ResourceMetrics()
		rmsExp := expected[q].ResourceMetrics()
		require.Equal(t, rmsExp.Len(), rmsAct.Len())
		for i := 0; i < rmsAct.Len(); i++ {
			rmAct := rmsAct.At(i)
			rmExp := rmsExp.At(i)

			// require equality of resource attributes
			require.Equal(t, rmExp.Resource().Attributes().Sort(), rmAct.Resource().Attributes().Sort())

			// require equality of IL metrics
			ilmsAct := rmAct.InstrumentationLibraryMetrics()
			ilmsExp := rmExp.InstrumentationLibraryMetrics()
			require.Equal(t, ilmsExp.Len(), ilmsAct.Len())
			for j := 0; j < ilmsAct.Len(); j++ {
				ilmAct := ilmsAct.At(j)
				ilmExp := ilmsExp.At(j)

				// require equality of metrics
				metricsAct := ilmAct.Metrics()
				metricsExp := ilmExp.Metrics()
				require.Equal(t, metricsExp.Len(), metricsAct.Len())

				// build a map of expected metrics
				metricsExpMap := make(map[string]pdata.Metric, metricsExp.Len())
				for k := 0; k < metricsExp.Len(); k++ {
					metricsExpMap[metricsExp.At(k).Name()] = metricsExp.At(k)
				}

				for k := 0; k < metricsAct.Len(); k++ {
					metricAct := metricsAct.At(k)
					metricExp := metricsExp.At(k)

					// require equality of descriptors
					require.Equal(t, metricExp.Name(), metricAct.Name())
					require.Equalf(t, metricExp.Description(), metricAct.Description(), "Metric %s", metricAct.Name())
					require.Equalf(t, metricExp.Unit(), metricAct.Unit(), "Metric %s", metricAct.Name())
					require.Equalf(t, metricExp.DataType(), metricAct.DataType(), "Metric %s", metricAct.Name())

					// require equality of aggregation info & data points
					switch ty := metricAct.DataType(); ty {
					case pdata.MetricDataTypeSum:
						require.Equal(t, metricAct.Sum().AggregationTemporality(), metricExp.Sum().AggregationTemporality(), "Metric %s", metricAct.Name())
						require.Equal(t, metricAct.Sum().IsMonotonic(), metricExp.Sum().IsMonotonic(), "Metric %s", metricAct.Name())
						requireEqualNumberDataPointSlice(t, metricAct.Name(), metricAct.Sum().DataPoints(), metricExp.Sum().DataPoints())
					case pdata.MetricDataTypeGauge:
						requireEqualNumberDataPointSlice(t, metricAct.Name(), metricAct.Gauge().DataPoints(), metricExp.Gauge().DataPoints())
					default:
						require.Fail(t, "unexpected metric type", t)
					}
				}
			}
		}
	}
}

func requireEqualNumberDataPointSlice(t *testing.T, metricName string, ddpsAct, ddpsExp pdata.NumberDataPointSlice) {
	require.Equalf(t, ddpsExp.Len(), ddpsAct.Len(), "Metric %s", metricName)

	// build a map of expected data points
	ddpsExpMap := make(map[string]pdata.NumberDataPoint, ddpsExp.Len())
	for k := 0; k < ddpsExp.Len(); k++ {
		ddpsExpMap[dataPointKey(metricName, ddpsExp.At(k))] = ddpsExp.At(k)
	}

	for l := 0; l < ddpsAct.Len(); l++ {
		ddpAct := ddpsAct.At(l)

		ddpExp, ok := ddpsExpMap[dataPointKey(metricName, ddpAct)]
		if !ok {
			require.Failf(t, fmt.Sprintf("no data point for %s", dataPointKey(metricName, ddpAct)), "Metric %s", metricName)
		}

		require.Equalf(t, ddpExp.Attributes().Sort(), ddpAct.Attributes().Sort(), "Metric %s", metricName)
		require.Equalf(t, ddpExp.StartTimestamp(), ddpAct.StartTimestamp(), "Metric %s", metricName)
		require.Equalf(t, ddpExp.Timestamp(), ddpAct.Timestamp(), "Metric %s", metricName)
		require.Equalf(t, ddpExp.Type(), ddpAct.Type(), "Metric %s", metricName)
		if ddpExp.Type() == pdata.MetricValueTypeDouble {
			require.InDeltaf(t, ddpExp.DoubleVal(), ddpAct.DoubleVal(), 0.00000001, "Metric %s", metricName)
		} else {
			require.Equalf(t, ddpExp.IntVal(), ddpAct.IntVal(), "Metric %s", metricName)
		}
	}
}

// dataPointKey returns a key representing the data point
func dataPointKey(metricName string, dataPoint pdata.NumberDataPoint) string {
	otherAttributesLen := dataPoint.Attributes().Len()

	idx, otherAttributes := 0, make([]string, otherAttributesLen)
	dataPoint.Attributes().Range(func(k string, v pdata.AttributeValue) bool {
		otherAttributes[idx] = fmt.Sprintf("%s=%v", k, v.AsString())
		idx++
		return true
	})
	// sort the slice so that we consider AttributeSets
	// the same regardless of order
	sort.StringSlice(otherAttributes).Sort()
	return metricName + "/" + dataPoint.StartTimestamp().String() + "-" + dataPoint.Timestamp().String() + "/" + strings.Join(otherAttributes, ";")
}
