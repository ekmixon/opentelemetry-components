package httpdreceiver

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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

func addToIntMetric(metric pdata.NumberDataPointSlice, attributes pdata.AttributeMap, value int64, ts pdata.Timestamp) {
	dataPoint := metric.AppendEmpty()
	dataPoint.SetTimestamp(ts)
	dataPoint.SetIntVal(value)
	if attributes.Len() > 0 {
		attributes.CopyTo(dataPoint.Attributes())
	}
}

func (r *httpdScraper) scrape(context.Context) (pdata.Metrics, error) {
	if r.httpClient == nil {
		return pdata.Metrics{}, errors.New("failed to connect to http client")
	}

	stats, err := r.GetStats()
	if err != nil {
		r.logger.Error("Failed to fetch httpd stats", zap.Error(err))
		return pdata.Metrics{}, err
	}

	rms := pdata.NewMetrics()
	ilm := rms.ResourceMetrics().AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty()
	ilm.InstrumentationLibrary().SetName("otelcol/httpd")
	now := pdata.NewTimestampFromTime(time.Now())

	uptime := initMetric(ilm.Metrics(), metadata.M.HttpdUptime).Sum().DataPoints()
	connections := initMetric(ilm.Metrics(), metadata.M.HttpdCurrentConnections).Gauge().DataPoints()
	workers := initMetric(ilm.Metrics(), metadata.M.HttpdWorkers).Gauge().DataPoints()
	requests := initMetric(ilm.Metrics(), metadata.M.HttpdRequests).Sum().DataPoints()
	traffic := initMetric(ilm.Metrics(), metadata.M.HttpdTraffic).Sum().DataPoints()
	scoreboard := initMetric(ilm.Metrics(), metadata.M.HttpdScoreboard).Gauge().DataPoints()

	u, err := url.Parse(r.cfg.Endpoint)
	if err != nil {
		r.logger.Error("Failed to find parse server name", zap.Error(err))
		return pdata.Metrics{}, err
	}
	serverName := u.Hostname()

	for metricKey, metricValue := range parseStats(stats) {
		attributes := pdata.NewAttributeMap()
		attributes.Insert(metadata.A.ServerName, pdata.NewAttributeValueString(serverName))

		switch metricKey {
		case "ServerUptimeSeconds":
			if i, ok := r.parseInt(metricKey, metricValue); ok {
				addToIntMetric(uptime, attributes, i, now)
			}
		case "ConnsTotal":
			if i, ok := r.parseInt(metricKey, metricValue); ok {
				addToIntMetric(connections, attributes, i, now)
			}
		case "BusyWorkers":
			if i, ok := r.parseInt(metricKey, metricValue); ok {
				attributes.Insert(metadata.A.WorkersState, pdata.NewAttributeValueString("busy"))
				addToIntMetric(workers, attributes, i, now)
			}
		case "IdleWorkers":
			if i, ok := r.parseInt(metricKey, metricValue); ok {
				attributes.Insert(metadata.A.WorkersState, pdata.NewAttributeValueString("idle"))
				addToIntMetric(workers, attributes, i, now)
			}
		case "Total Accesses":
			if i, ok := r.parseInt(metricKey, metricValue); ok {
				addToIntMetric(requests, attributes, i, now)
			}
		case "Total kBytes":
			if i, ok := r.parseInt(metricKey, metricValue); ok {
				bytes := kbytesToBytes(i)
				addToIntMetric(traffic, attributes, bytes, now)
			}
		case "Scoreboard":
			scoreboardMap := parseScoreboard(metricValue)
			for identifier, score := range scoreboardMap {
				attributes.Upsert(metadata.A.ScoreboardState, pdata.NewAttributeValueString(identifier))
				addToIntMetric(scoreboard, attributes, score, now)
			}
		}
	}

	return rms, nil
}

// GetStats collects metric stats by making a get request at an endpoint.
func (r *httpdScraper) GetStats() (string, error) {
	url := fmt.Sprintf("%s%s", r.cfg.Endpoint, "/server-status?auto")
	resp, err := r.httpClient.Get(url)
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

// kbytesToBytes converts 1 Kilobyte to 1024 bytes.
func kbytesToBytes(i int64) int64 {
	return 1024 * i
}
