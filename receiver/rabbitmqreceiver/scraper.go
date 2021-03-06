package rabbitmqreceiver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/observiq/opentelemetry-components/receiver/rabbitmqreceiver/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"
)

type rabbitmqScraper struct {
	httpClient *http.Client
	logger     *zap.Logger
	cfg        *Config
}

func newRabbitMQScraper(
	logger *zap.Logger,
	cfg *Config,
) (*rabbitmqScraper, error) {
	return &rabbitmqScraper{
		logger: logger,
		cfg:    cfg,
	}, nil
}

func (r *rabbitmqScraper) start(_ context.Context, host component.Host) error {
	httpClient, err := r.cfg.ToClient(host.GetExtensions())
	if err != nil {
		return err
	}
	r.httpClient = httpClient
	return nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (r *rabbitmqScraper) scrape(context.Context) (pdata.Metrics, error) {
	req, err := http.NewRequest("GET", r.cfg.Endpoint+"/api/queues", nil)
	if err != nil {
		return pdata.Metrics{}, err
	}

	req.Header.Add("Authorization", "Basic "+basicAuth(r.cfg.Username, r.cfg.Password))
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return pdata.Metrics{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return pdata.Metrics{}, err
	}

	var bodyParsed []interface{}

	err = json.Unmarshal(body, &bodyParsed)
	if err != nil {
		return pdata.Metrics{}, err
	}
	rms := pdata.NewMetrics()
	ilm := rms.ResourceMetrics().AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty()
	ilm.InstrumentationLibrary().SetName("otelcol/rabbitmq")
	now := pdata.NewTimestampFromTime(time.Now())

	publishRateMetric := initMetric(ilm.Metrics(), metadata.M.RabbitmqPublishRate).Gauge().DataPoints()
	deliveryRateMetric := initMetric(ilm.Metrics(), metadata.M.RabbitmqDeliveryRate).Gauge().DataPoints()
	consumersMetric := initMetric(ilm.Metrics(), metadata.M.RabbitmqConsumers).Gauge().DataPoints()
	numMessagesMetric := initMetric(ilm.Metrics(), metadata.M.RabbitmqNumMessages).Gauge().DataPoints()

	for _, v := range bodyParsed {
		queue, ok := v.(map[string]interface{})
		if !ok {
			r.logger.Info("rabbitMQ api response format did not meet expectations")
			break
		}
		attributes := pdata.NewAttributeMap()

		queueName, ok := queue["name"].(string)
		if !ok {
			r.logger.Info("could not parse queue name from body")
			break
		}
		attributes.Upsert(metadata.A.Queue, pdata.NewAttributeValueString(queueName))

		val, err := getValFromBody([]string{"message_stats", "publish_details", "rate"}, queue)
		if err != nil {
			r.logger.Info(
				err.Error(),
				zap.String("metric", "publish_rate"),
			)
		} else {
			addToDoubleMetric(publishRateMetric, attributes, val, now)
		}

		val, err = getValFromBody([]string{"message_stats", "deliver_details", "rate"}, queue)
		if err != nil {
			r.logger.Info(
				err.Error(),
				zap.String("metric", "delivery_rate"),
			)
		} else {
			addToDoubleMetric(deliveryRateMetric, attributes, val, now)
		}

		val, err = getValFromBody([]string{"consumers"}, queue)
		if err != nil {
			r.logger.Info(
				err.Error(),
				zap.String("metric", "consumers"),
			)
		} else {
			addToDoubleMetric(consumersMetric, attributes, val, now)
		}

		val, err = getValFromBody([]string{"messages"}, queue)
		if err != nil {
			r.logger.Info(
				err.Error(),
				zap.String("metric", "num_messages state:total"),
			)
		} else {
			attributes.Upsert(metadata.A.State, pdata.NewAttributeValueString("total"))
			addToDoubleMetric(numMessagesMetric, attributes, val, now)
		}

		val, err = getValFromBody([]string{"messages_unacknowledged"}, queue)
		if err != nil {
			r.logger.Info(
				err.Error(),
				zap.String("metric", "num_messages state:unacknowledged"),
			)
		} else {
			attributes.Upsert(metadata.A.State, pdata.NewAttributeValueString("unacknowledged"))
			addToDoubleMetric(numMessagesMetric, attributes, val, now)
		}

		val, err = getValFromBody([]string{"messages_ready"}, queue)
		if err != nil {
			r.logger.Info(
				err.Error(),
				zap.String("metric", "num_messages state:ready"),
			)
		} else {
			attributes.Upsert(metadata.A.State, pdata.NewAttributeValueString("ready"))
			addToDoubleMetric(numMessagesMetric, attributes, val, now)
		}
	}

	return rms, nil
}

func getValFromBody(keys []string, body map[string]interface{}) (float64, error) {
	var currentValue interface{} = body

	for _, key := range keys {
		currentBody, ok := currentValue.(map[string]interface{})
		if !ok {
			return 0, fmt.Errorf("could not find key in body")
		}

		currentValue, ok = currentBody[key]
		if !ok {
			return 0, fmt.Errorf("could not find key in body")
		}
	}
	floatVal, ok := parseFloat(currentValue)
	if !ok {
		return 0, fmt.Errorf("could not parse value as float")
	}
	return floatVal, nil
}

// parseFloat converts string to float64.
func parseFloat(value interface{}) (float64, bool) {
	switch f := value.(type) {
	case float64:
		return f, true
	case int64:
		return float64(f), true
	case float32:
		return float64(f), true
	case int32:
		return float64(f), true
	case string:
		fConv, err := strconv.ParseFloat(f, 64)
		if err != nil {
			return 0, false
		}
		return fConv, true
	}
	return 0, false
}

func initMetric(ms pdata.MetricSlice, mi metadata.MetricIntf) pdata.Metric {
	m := ms.AppendEmpty()
	mi.Init(m)
	return m
}

func addToDoubleMetric(metric pdata.NumberDataPointSlice, attributes pdata.AttributeMap, value float64, ts pdata.Timestamp) {
	dataPoint := metric.AppendEmpty()
	dataPoint.SetTimestamp(ts)
	dataPoint.SetDoubleVal(value)
	if attributes.Len() > 0 {
		attributes.CopyTo(dataPoint.Attributes())
	}
}
