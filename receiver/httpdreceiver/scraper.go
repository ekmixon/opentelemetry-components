package httpdreceiver

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"

	"github.com/observiq/opentelemetry-components/receiver/httpdreceiver/internal/metadata"
)

type httpdScraper struct {
	logger     *zap.Logger
	cfg        *Config
	httpClient *http.Client
}

func newHttpdScraper(
	logger *zap.Logger,
	cfg *Config,
) *httpdScraper {
	return &httpdScraper{
		logger: logger,
		cfg:    cfg,
	}
}

func (r *httpdScraper) start(_ context.Context, host component.Host) error {
	httpClient, err := r.cfg.ToClient(host.GetExtensions())
	if err != nil {
		return err
	}
	r.httpClient = httpClient
	return nil
}

func initMetric(ms pdata.MetricSlice, mi metadata.MetricIntf) pdata.Metric {
	m := ms.AppendEmpty()
	mi.Init(m)
	return m
}

func addToDoubleMetric(metric pdata.NumberDataPointSlice, labels pdata.StringMap, value float64, ts pdata.Timestamp) {
	dataPoint := metric.AppendEmpty()
	dataPoint.SetTimestamp(ts)
	dataPoint.SetDoubleVal(value)
	if labels.Len() > 0 {
		labels.CopyTo(dataPoint.LabelsMap())
	}
}

func addToIntMetric(metric pdata.NumberDataPointSlice, labels pdata.StringMap, value int64, ts pdata.Timestamp) {
	dataPoint := metric.AppendEmpty()
	dataPoint.SetTimestamp(ts)
	dataPoint.SetIntVal(value)
	if labels.Len() > 0 {
		labels.CopyTo(dataPoint.LabelsMap())
	}
}

func (r *httpdScraper) scrape(context.Context) (pdata.ResourceMetricsSlice, error) {
	if r.httpClient == nil {
		return pdata.ResourceMetricsSlice{}, errors.New("failed to connect to http client")
	}

	stats, err := r.GetStats()
	if err != nil {
		r.logger.Error("Failed to fetch httpd stats", zap.Error(err))
		return pdata.ResourceMetricsSlice{}, err
	}

	metrics := pdata.NewMetrics()
	ilm := metrics.ResourceMetrics().AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty()
	ilm.InstrumentationLibrary().SetName("otel/httpd")
	now := pdata.TimestampFromTime(time.Now())

	uptime := initMetric(ilm.Metrics(), metadata.M.HttpdUptime).Sum().DataPoints()
	connections := initMetric(ilm.Metrics(), metadata.M.HttpdCurrentConnections).Gauge().DataPoints()
	workers := initMetric(ilm.Metrics(), metadata.M.HttpdWorkers).Gauge().DataPoints()
	requests := initMetric(ilm.Metrics(), metadata.M.HttpdRequests).Gauge().DataPoints()
	bytes := initMetric(ilm.Metrics(), metadata.M.HttpdBytes).Gauge().DataPoints()
	traffic := initMetric(ilm.Metrics(), metadata.M.HttpdTraffic).Sum().DataPoints()
	scoreboard := initMetric(ilm.Metrics(), metadata.M.HttpdScoreboard).Gauge().DataPoints()

	for metricKey, metricValue := range parseStats(stats) {
		labels := pdata.NewStringMap()
		switch metricKey {
		case "ServerUptimeSeconds":
			if i, ok := r.parseInt(metricKey, metricValue); ok {
				addToIntMetric(uptime, labels, i, now)
			}
		case "ConnsTotal":
			if i, ok := r.parseInt(metricKey, metricValue); ok {
				addToIntMetric(connections, labels, i, now)
			}
		case "BusyWorkers":
			if i, ok := r.parseInt(metricKey, metricValue); ok {
				labels.Insert(metadata.L.WorkersState, "busy")
				addToIntMetric(workers, labels, i, now)
			}
		case "IdleWorkers":
			if i, ok := r.parseInt(metricKey, metricValue); ok {
				labels.Insert(metadata.L.WorkersState, "idle")
				addToIntMetric(workers, labels, i, now)
			}
		case "ReqPerSec":
			if f, ok := r.parseFloat(metricKey, metricValue); ok {
				addToDoubleMetric(requests, labels, f, now)
			}
		case "BytesPerSec":
			if f, ok := r.parseFloat(metricKey, metricValue); ok {
				addToDoubleMetric(bytes, labels, f, now)
			}
		case "Total Accesses":
			if i, ok := r.parseInt(metricKey, metricValue); ok {
				addToIntMetric(traffic, labels, i, now)
			}
		case "Scoreboard":
			scoreboardMap := parseScoreboard(metricValue)
			for identifier, score := range scoreboardMap {
				labels := pdata.NewStringMap()
				labels.Insert(metadata.L.ScoreboardState, identifier)
				addToIntMetric(scoreboard, labels, score, now)
			}
		}
	}

	return metrics.ResourceMetrics(), nil
}

// GetStats collects metric stats by making a get request at an endpoint.
func (r *httpdScraper) GetStats() (string, error) {
	resp, err := r.httpClient.Get(r.cfg.HTTPClientSettings.Endpoint)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// parseStats converts a response body key:values into a map.
func parseStats(resp string) map[string]string {
	metrics := make(map[string]string)

	fields := strings.Split(resp, "\n")
	for _, field := range fields {
		index := strings.Index(field, ": ")
		if index == -1 {
			continue
		}
		metrics[field[:index]] = field[index+2:]
	}
	return metrics
}

// parseFloat converts string to float64.
func (r *httpdScraper) parseFloat(key, value string) (float64, bool) {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		r.logInvalid("float", key, value)
		return 0, false
	}
	return f, true
}

// parseInt converts string to int64.
func (r *httpdScraper) parseInt(key, value string) (int64, bool) {
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		r.logInvalid("int", key, value)
		return 0, false
	}
	return i, true
}

func (r *httpdScraper) logInvalid(expectedType, key, value string) {
	r.logger.Info(
		"invalid value",
		zap.String("expectedType", expectedType),
		zap.String("key", key),
		zap.String("value", value),
	)
}

// parseScoreboard quantifies the symbolic mapping of the scoreboard.
func parseScoreboard(values string) map[string]int64 {
	scoreboard := map[string]int64{
		"waiting":      0,
		"starting":     0,
		"reading":      0,
		"sending":      0,
		"keepalive":    0,
		"dnslookup":    0,
		"closing":      0,
		"logging":      0,
		"finishing":    0,
		"idle_cleanup": 0,
		"open":         0,
	}

	for _, char := range values {
		switch string(char) {
		case "_":
			scoreboard["waiting"] += 1
		case "S":
			scoreboard["starting"] += 1
		case "R":
			scoreboard["reading"] += 1
		case "W":
			scoreboard["sending"] += 1
		case "K":
			scoreboard["keepalive"] += 1
		case "D":
			scoreboard["dnslookup"] += 1
		case "C":
			scoreboard["closing"] += 1
		case "L":
			scoreboard["logging"] += 1
		case "G":
			scoreboard["finishing"] += 1
		case "I":
			scoreboard["idle_cleanup"] += 1
		case ".":
			scoreboard["open"] += 1
		}
	}
	return scoreboard
}
