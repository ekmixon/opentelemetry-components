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
	httpClient *http.Client

	logger *zap.Logger
	cfg    *Config
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

func addToMetric(metric pdata.DoubleDataPointSlice, labels pdata.StringMap, value float64, ts pdata.Timestamp) {
	dataPoint := metric.AppendEmpty()
	dataPoint.SetTimestamp(ts)
	dataPoint.SetValue(value)
	if labels.Len() > 0 {
		labels.CopyTo(dataPoint.LabelsMap())
	}
}

func addToIntMetric(metric pdata.IntDataPointSlice, labels pdata.StringMap, value int64, ts pdata.Timestamp) {
	dataPoint := metric.AppendEmpty()
	dataPoint.SetTimestamp(ts)
	dataPoint.SetValue(value)
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

	uptime := initMetric(ilm.Metrics(), metadata.M.HttpdUptime).IntSum().DataPoints()
	connections := initMetric(ilm.Metrics(), metadata.M.HttpdCurrentConnections).IntGauge().DataPoints()
	workers := initMetric(ilm.Metrics(), metadata.M.HttpdWorkers).IntGauge().DataPoints()
	requests := initMetric(ilm.Metrics(), metadata.M.HttpdRequests).Gauge().DataPoints()
	bytes := initMetric(ilm.Metrics(), metadata.M.HttpdBytes).Gauge().DataPoints()
	traffic := initMetric(ilm.Metrics(), metadata.M.HttpdTraffic).IntSum().DataPoints()
	scoreboard := initMetric(ilm.Metrics(), metadata.M.HttpdScoreboard).IntGauge().DataPoints()

	for metricKey, metricValue := range parseStats(stats) {
		labels := pdata.NewStringMap()
		switch metricKey {
		case "ServerUptimeSeconds":
			addToIntMetric(uptime, labels, parseInt(metricValue), now)
		case "ConnsTotal":
			addToIntMetric(connections, labels, parseInt(metricValue), now)
		case "BusyWorkers":
			labels.Insert(metadata.L.WorkersState, "busy")
			addToIntMetric(workers, labels, parseInt(metricValue), now)
		case "IdleWorkers":
			labels.Insert(metadata.L.WorkersState, "idle")
			addToIntMetric(workers, labels, parseInt(metricValue), now)
		case "ReqPerSec":
			addToMetric(requests, labels, parseFloat(metricValue), now)
		case "BytesPerSec":
			addToMetric(bytes, labels, parseFloat(metricValue), now)
		case "Total Accesses":
			addToIntMetric(traffic, labels, parseInt(metricValue), now)
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
func parseFloat(value string) float64 {
	f, _ := strconv.ParseFloat(value, 64)
	return f
}

// parseInt converts string to int64.
func parseInt(value string) int64 {
	i, _ := strconv.ParseInt(value, 10, 64)
	return i
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
