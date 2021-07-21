package rabbitmqreceiver

import (
	"context"
	"encoding/base64"
	"encoding/json"
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
) *rabbitmqScraper {
	return &rabbitmqScraper{
		logger: logger,
		cfg:    cfg,
	}
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

func getValFromBody(keys []string, body map[string]interface{}) (interface{}, bool) {
	var currentValue interface{} = body

	for _, key := range keys {
		currentBody, ok := currentValue.(map[string]interface{})
		if !ok {
			return nil, false
		}

		currentValue, ok = currentBody[key]
		if !ok {
			return nil, false
		}
	}

	return currentValue, true
}

func (r *rabbitmqScraper) scrape(context.Context) (pdata.ResourceMetricsSlice, error) {
	req, err := http.NewRequest("GET", r.cfg.Endpoint+"/api/queues", nil)
	if err != nil {
		return pdata.ResourceMetricsSlice{}, err
	}

	req.Header.Add("Authorization", "Basic "+basicAuth(r.cfg.Username, r.cfg.Password))
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return pdata.ResourceMetricsSlice{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return pdata.ResourceMetricsSlice{}, err
	}

	var bodyParsed []interface{}

	err = json.Unmarshal(body, &bodyParsed)
	if err != nil {
		return pdata.ResourceMetricsSlice{}, err
	}
	metrics := pdata.NewMetrics()
	ilm := metrics.ResourceMetrics().AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty()
	ilm.InstrumentationLibrary().SetName("otel/rabbitmq")
	now := pdata.TimestampFromTime(time.Now())

	publishRateMetric := initMetric(ilm.Metrics(), metadata.M.RabbitmqPublishRate).Gauge().DataPoints()
	deliveryRateMetric := initMetric(ilm.Metrics(), metadata.M.RabbitmqDeliveryRate).Gauge().DataPoints()
	consumersMetric := initMetric(ilm.Metrics(), metadata.M.RabbitmqConsumers).Gauge().DataPoints()
	numMessagesMetric := initMetric(ilm.Metrics(), metadata.M.RabbitmqNumMessages).Gauge().DataPoints()

	for _, v := range bodyParsed {
		queue, _ := v.(map[string]interface{})
		labels := pdata.NewStringMap()

		queueName, ok := getValFromBody([]string{"name"}, queue)
		r.isOK(ok, "could not get queue name from body")
		queueNameStr, ok := queueName.(string)
		r.isOK(ok, "could not parse queue name as string")
		labels.Upsert(metadata.L.Queue, queueNameStr)

		publishRateInter, ok := getValFromBody([]string{"message_stats", "publish_details", "rate"}, queue)
		r.isOK(ok, "could not get publish_rate from body")
		publishRateVal, ok := r.parseFloat("publish_rate", publishRateInter)
		r.isOK(ok, "could not parse publish_rate as a float")
		addToMetric(publishRateMetric, labels, publishRateVal, now)

		deliveryRateInter, ok := getValFromBody([]string{"message_stats", "deliver_details", "rate"}, queue)
		r.isOK(ok, "could not get delivery_rate from body")
		deliveryRateVal, ok := r.parseFloat("delivery_rate", deliveryRateInter)
		r.isOK(ok, "could not parse delivery_rate as a float")
		addToMetric(deliveryRateMetric, labels, deliveryRateVal, now)

		consumersInter, ok := getValFromBody([]string{"consumers"}, queue)
		r.isOK(ok, "could not get consumers from body")
		consumersVal, ok := r.parseFloat("consumers", consumersInter)
		r.isOK(ok, "could not parse consumers as a float")
		addToMetric(consumersMetric, labels, consumersVal, now)

		numMessagesInter, ok := getValFromBody([]string{"messages"}, queue)
		r.isOK(ok, "could not get messages from body")
		numMessagesVal, ok := r.parseFloat("num_messages - messages", numMessagesInter)
		labels.Upsert(metadata.L.State, "total")
		r.isOK(ok, "could not parse messages as a float")
		addToMetric(numMessagesMetric, labels, numMessagesVal, now)

		numMessagesInter, ok = getValFromBody([]string{"messages_unacknowledged"}, queue)
		r.isOK(ok, "could not get messages_unacknowledged from body")
		numMessagesVal, ok = r.parseFloat("num_messages - messages_unacknowledged", numMessagesInter)
		r.isOK(ok, "could not parse messages_unacknowledged as a float")
		labels.Upsert(metadata.L.State, "unacknowledged")
		addToMetric(numMessagesMetric, labels, numMessagesVal, now)

		numMessagesInter, ok = getValFromBody([]string{"messages_ready"}, queue)
		r.isOK(ok, "could not get messages_ready from body")
		numMessagesVal, ok = r.parseFloat("num_messages - ready", numMessagesInter)
		r.isOK(ok, "could not parse messages_ready as a float")
		labels.Upsert(metadata.L.State, "ready")
		addToMetric(numMessagesMetric, labels, numMessagesVal, now)
	}

	return metrics.ResourceMetrics(), nil
}

// parseFloat converts string to float64.
func (r *rabbitmqScraper) parseFloat(key string, value interface{}) (float64, bool) {
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
			r.logInvalid("float", key, f)
			return 0, false
		}
		return fConv, true
	}
	return 0, false
}

func (r *rabbitmqScraper) logInvalid(expectedType, key, value string) {
	r.logger.Info(
		"invalid value",
		zap.String("expectedType", expectedType),
		zap.String("key", key),
		zap.String("value", value),
	)
}

func (r *rabbitmqScraper) isOK(ok bool, message string) {
	if !ok {
		r.logger.Info(
			message,
		)
	}
}

func initMetric(ms pdata.MetricSlice, mi metadata.MetricIntf) pdata.Metric {
	m := ms.AppendEmpty()
	mi.Init(m)
	return m
}

func addToMetric(metric pdata.DoubleDataPointSlice, labels pdata.StringMap, value float64, ts pdata.Timestamp) {
	dataPoint := metric.AppendEmpty()
	dataPoint.SetTimestamp(ts)
	dataPoint.SetValue(value)
	if labels.Len() > 0 {
		labels.CopyTo(dataPoint.LabelsMap())
	}
}
