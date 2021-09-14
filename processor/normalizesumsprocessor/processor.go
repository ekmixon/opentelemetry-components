package normalizesumsprocessor

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"
)

type NormalizeSumsProcessor struct {
	logger     *zap.Logger
	transforms []Transform

	historyMux sync.RWMutex
	history    map[string]*startPoint
}

type startPoint struct {
	dataType        pdata.MetricDataType
	numberDataPoint *pdata.NumberDataPoint
	lastDoubleValue float64
}

func newNormalizeSumsProcessor(logger *zap.Logger, transforms []Transform) *NormalizeSumsProcessor {
	return &NormalizeSumsProcessor{
		logger:     logger,
		transforms: transforms,
		history:    make(map[string]*startPoint),
	}
}

// Start is invoked during service startup.
func (nsp *NormalizeSumsProcessor) Start(context.Context, component.Host) error {
	return nil
}

// Shutdown is invoked during service shutdown.
func (nsp *NormalizeSumsProcessor) Shutdown(context.Context) error {
	return nil
}

// ProcessMetrics implements the MProcessor interface.
func (nsp *NormalizeSumsProcessor) ProcessMetrics(ctx context.Context, metrics pdata.Metrics) (pdata.Metrics, error) {
	var errors []error

	for i := 0; i < metrics.ResourceMetrics().Len(); i++ {
		rms := metrics.ResourceMetrics().At(i)
		processingErrors := nsp.transformMetrics(rms)
		errors = append(errors, processingErrors...)
	}

	if len(errors) > 0 {
		return metrics, consumererror.Combine(errors)
	}

	return metrics, nil
}

func (nsp *NormalizeSumsProcessor) transformMetrics(rms pdata.ResourceMetrics) []error {
	var errors []error

	ilms := rms.InstrumentationLibraryMetrics()
	for j := 0; j < ilms.Len(); j++ {
		ilm := ilms.At(j).Metrics()
		newSlice := pdata.NewMetricSlice()
		for k := 0; k < ilm.Len(); k++ {
			metric := ilm.At(k)
			if shouldTransform, transform := nsp.shouldTransformMetric(metric); shouldTransform {
				keepMetric, err := nsp.processMetric(rms.Resource(), metric)
				if err != nil {
					errors = append(errors, err)
				}
				if keepMetric {
					newMetric := newSlice.AppendEmpty()
					metric.CopyTo(newMetric)
					if transform.NewName != "" {
						newMetric.SetName(transform.NewName)
					}
				}
			} else {
				newMetric := newSlice.AppendEmpty()
				metric.CopyTo(newMetric)
			}
		}

		newSlice.CopyTo(ilm)
	}

	return errors
}

func (nsp *NormalizeSumsProcessor) shouldTransformMetric(metric pdata.Metric) (bool, *Transform) {
	// Only consider Sums
	if metric.DataType() != pdata.MetricDataTypeSum {
		return false, nil
	}

	// If transforms is empty, transform all Sums
	if nsp.transforms == nil {
		t := Transform{MetricName: metric.Name()}
		return true, &t
	}

	// Check through the list of transforms for the named metric
	for _, transform := range nsp.transforms {
		if transform.MetricName == metric.Name() {
			return true, &transform
		}
	}

	return false, nil
}

func (nsp *NormalizeSumsProcessor) processMetric(resource pdata.Resource, metric pdata.Metric) (bool, error) {
	switch t := metric.DataType(); t {
	case pdata.MetricDataTypeSum:
		return nsp.processSumMetric(resource, metric) > 0, nil
	default:
		return false, fmt.Errorf("data type not supported %s", t)
	}
}

func (nsp *NormalizeSumsProcessor) processSumMetric(resource pdata.Resource, metric pdata.Metric) int {
	dps := metric.Sum().DataPoints()
	for i := 0; i < dps.Len(); {
		reportData := nsp.processSumDataPoint(dps.At(i), resource, metric)

		if !reportData {
			removeAt(dps, i)
			continue
		}
		i++
	}

	return dps.Len()
}

func (nsp *NormalizeSumsProcessor) processSumDataPoint(dp pdata.NumberDataPoint, resource pdata.Resource, metric pdata.Metric) bool {
	metricIdentifier := dataPointIdentifier(resource, metric, dp.LabelsMap())

	nsp.historyMux.RLock()
	start := nsp.history[metricIdentifier]
	nsp.historyMux.RUnlock()
	// If this is the first time we've observed this unique metric,
	// record it as the start point and do not report this data point
	if start == nil {
		dps := metric.Sum().DataPoints()
		newDP := pdata.NewNumberDataPoint()
		dps.At(0).CopyTo(newDP)

		newStart := startPoint{
			dataType:        pdata.MetricDataTypeSum,
			numberDataPoint: &newDP,
			lastDoubleValue: newDP.Value(),
		}
		nsp.historyMux.Lock()
		nsp.history[metricIdentifier] = &newStart
		nsp.historyMux.Unlock()

		return false
	}

	// If this data is older than the start point, we can't
	// meaningfully report this point
	if dp.Timestamp() <= start.numberDataPoint.Timestamp() {
		return false
	}

	// If data has rolled over or the counter has been restarted for
	// any other reason, grab a new start point and do not report this data
	if dp.Value() < start.lastDoubleValue {
		dp.CopyTo(*start.numberDataPoint)
		start.lastDoubleValue = dp.Value()

		return false
	}

	start.lastDoubleValue = dp.Value()
	dp.SetDoubleVal(dp.Value() - start.numberDataPoint.Value())
	dp.SetStartTimestamp(start.numberDataPoint.Timestamp())

	return true
}

func dataPointIdentifier(resource pdata.Resource, metric pdata.Metric, labels pdata.StringMap) string {
	var b strings.Builder

	// Resource identifiers
	resource.Attributes().Sort().Range(func(k string, v pdata.AttributeValue) bool {
		fmt.Fprintf(&b, "%s=", k)
		addAttributeToIdentityBuilder(&b, v)
		b.WriteString("|")
		return true
	})

	// Metric identifiers
	fmt.Fprintf(&b, " - %s", metric.Name())
	labels.Sort().Range(func(k, v string) bool {
		fmt.Fprintf(&b, " %s=%s", k, v)
		return true
	})
	return b.String()
}

func addAttributeToIdentityBuilder(b *strings.Builder, v pdata.AttributeValue) {
	switch v.Type() {
	case pdata.AttributeValueTypeArray:
		b.WriteString("[")
		arr := v.ArrayVal()
		for i := 0; i < arr.Len(); i++ {
			addAttributeToIdentityBuilder(b, arr.At(i))
			b.WriteString(",")
		}
		b.WriteString("]")
	case pdata.AttributeValueTypeBool:
		fmt.Fprintf(b, "%t", v.BoolVal())
	case pdata.AttributeValueTypeDouble:
		// TODO - Double attribute values could be problematic for use in
		// forming an identify due to floating point math. Consider how to best
		// handle this case
		fmt.Fprintf(b, "%f", v.DoubleVal())
	case pdata.AttributeValueTypeInt:
		fmt.Fprintf(b, "%d", v.IntVal())
	case pdata.AttributeValueTypeMap:
		b.WriteString("{")
		v.MapVal().Sort().Range(func(k string, mapVal pdata.AttributeValue) bool {
			fmt.Fprintf(b, "%s:", k)
			addAttributeToIdentityBuilder(b, mapVal)
			b.WriteString(",")
			return true
		})
		b.WriteString("}")
	case pdata.AttributeValueTypeNull:
		b.WriteString("NULL")
	case pdata.AttributeValueTypeString:
		fmt.Fprintf(b, "'%s'", v.StringVal())
	}
}

func removeAt(slice pdata.NumberDataPointSlice, idx int) {
	newSlice := pdata.NewNumberDataPointSlice()
	j := 0
	for i := 0; i < slice.Len(); i++ {
		if i != idx {
			dp := newSlice.AppendEmpty()
			slice.At(i).CopyTo(dp)
			j++
		}
	}
	newSlice.CopyTo(slice)
}
