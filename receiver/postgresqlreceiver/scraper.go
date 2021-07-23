package postgresqlreceiver

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"

	"github.com/observiq/opentelemetry-components/receiver/postgresqlreceiver/internal/metadata"
)

type postgreSQLScraper struct {
	client   client
	stopOnce sync.Once

	logger *zap.Logger
	config *Config
}

func newPostgreSQLScraper(
	logger *zap.Logger,
	config *Config,
) *postgreSQLScraper {
	return &postgreSQLScraper{
		logger: logger,
		config: config,
	}
}

// start starts the scraper by initializing the db client connection.
func (p *postgreSQLScraper) start(_ context.Context, host component.Host) error {
	client, err := newPostgreSQLClient(postgreSQLConfig{
		username:     p.config.Username,
		password:     p.config.Password,
		databaseName: p.config.Database,
		endpoint:     p.config.Endpoint,
	})
	if err != nil {
		return err
	}
	p.client = client

	return nil
}

// shutdown closes open connections.
func (p *postgreSQLScraper) shutdown(context.Context) error {
	var err error
	p.stopOnce.Do(func() {
		err = p.client.Close()
	})
	return err
}

// initMetric initializes a metric with a metadata label.
func initMetric(ms pdata.MetricSlice, mi metadata.MetricIntf) pdata.Metric {
	m := ms.AppendEmpty()
	mi.Init(m)
	return m
}

// addToMetric adds and labels a double gauge datapoint to a metricslice.
func addToMetric(metric pdata.DoubleDataPointSlice, labels pdata.StringMap, value float64, ts pdata.Timestamp) {
	dataPoint := metric.AppendEmpty()
	dataPoint.SetTimestamp(ts)
	dataPoint.SetValue(value)
	if labels.Len() > 0 {
		labels.CopyTo(dataPoint.LabelsMap())
	}
}

// addToIntMetric adds and labels a int sum datapoint to metricslice.
func addToIntMetric(metric pdata.IntDataPointSlice, labels pdata.StringMap, value int64, ts pdata.Timestamp) {
	dataPoint := metric.AppendEmpty()
	dataPoint.SetTimestamp(ts)
	dataPoint.SetValue(value)
	if labels.Len() > 0 {
		labels.CopyTo(dataPoint.LabelsMap())
	}
}

// scrape scrapes the metric stats, transforms them and labels them into a metric slices.
func (p *postgreSQLScraper) scrape(context.Context) (pdata.ResourceMetricsSlice, error) {

	if p.client == nil {
		return pdata.ResourceMetricsSlice{}, errors.New("failed to connect to http client")
	}

	// metric initialization
	rms := pdata.NewResourceMetricsSlice()
	ilm := rms.AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty()
	ilm.InstrumentationLibrary().SetName("otel/postgresql")
	now := pdata.TimestampFromTime(time.Now())

	blocks_read := initMetric(ilm.Metrics(), metadata.M.PostgresqlBlocksRead).IntSum().DataPoints()
	commits := initMetric(ilm.Metrics(), metadata.M.PostgresqlCommits).IntSum().DataPoints()
	databaseSize := initMetric(ilm.Metrics(), metadata.M.PostgresqlDbSize).Gauge().DataPoints()
	backends := initMetric(ilm.Metrics(), metadata.M.PostgresqlBackends).Gauge().DataPoints()
	databaseRows := initMetric(ilm.Metrics(), metadata.M.PostgresqlRows).Gauge().DataPoints()
	operations := initMetric(ilm.Metrics(), metadata.M.PostgresqlOperations).IntSum().DataPoints()
	rollbacks := initMetric(ilm.Metrics(), metadata.M.PostgresqlRollbacks).Gauge().DataPoints()

	// blocks read query
	blocksReadMetric, err := p.client.getBlocksRead()
	if err != nil {
		p.logger.Error("Failed to fetch blocks read", zap.Error(err))
	} else {
		for k, v := range blocksReadMetric.stats {
			labels := pdata.NewStringMap()
			if i, ok := p.parseInt(k, v); ok {
				labels.Insert(metadata.L.Database, blocksReadMetric.database)
				labels.Insert(metadata.L.Table, blocksReadMetric.table)
				labels.Insert(metadata.L.Source, k)
				addToIntMetric(blocks_read, labels, i, now)
			}
		}
	}

	// blocks read by table
	blocksReadByTableMetrics, err := p.client.getBlocksReadByTable()
	if err != nil {
		p.logger.Error("Failed to fetch blocks read by table", zap.Error(err))
	} else {
		for _, table := range blocksReadByTableMetrics {
			for k, v := range table.stats {
				if i, ok := p.parseInt(k, v); ok {
					labels := pdata.NewStringMap()
					labels.Insert(metadata.L.Database, table.database)
					labels.Insert(metadata.L.Table, table.table)
					labels.Insert(metadata.L.Source, k)
					addToIntMetric(blocks_read, labels, i, now)
				}
			}
		}
	}

	// commits
	commitsMetric, err := p.client.getCommits()
	if err != nil {
		p.logger.Error("Failed to fetch commits", zap.Error(err))
	} else {
		for k, v := range commitsMetric.stats {
			labels := pdata.NewStringMap()
			if i, ok := p.parseInt(k, v); ok {
				labels.Insert(metadata.L.Database, commitsMetric.database)
				addToIntMetric(commits, labels, i, now)
			}
		}
	}

	// database size
	databaseSizeMetric, err := p.client.getDatabaseSize()
	if err != nil {
		p.logger.Error("Failed to fetch database size", zap.Error(err))
	} else {
		for k, v := range databaseSizeMetric.stats {
			labels := pdata.NewStringMap()
			if f, ok := p.parseFloat(k, v); ok {
				labels.Insert(metadata.L.Database, databaseSizeMetric.database)
				addToMetric(databaseSize, labels, f, now)
			}
		}
	}

	// backends
	backendsMetric, err := p.client.getBackends()
	if err != nil {
		p.logger.Error("Failed to fetch backends", zap.Error(err))
	} else {
		for k, v := range backendsMetric.stats {
			labels := pdata.NewStringMap()
			if f, ok := p.parseFloat(k, v); ok {
				labels.Insert(metadata.L.Database, backendsMetric.database)
				addToMetric(backends, labels, f, now)
			}
		}
	}

	// database rows query
	databaseRowsMetric, err := p.client.getDatabaseRows()
	if err != nil {
		p.logger.Error("Failed to fetch database rows", zap.Error(err))
	} else {
		for k, v := range databaseRowsMetric.stats {
			labels := pdata.NewStringMap()
			if f, ok := p.parseFloat(k, v); ok {
				labels.Insert(metadata.L.Database, databaseRowsMetric.database)
				labels.Insert(metadata.L.Table, databaseRowsMetric.table)
				labels.Insert(metadata.L.State, k)
				addToMetric(databaseRows, labels, f, now)
			}
		}
	}

	// database rows by table
	databaseRowsByTableMetrics, err := p.client.getDatabaseRowsByTable()
	if err != nil {
		p.logger.Error("Failed to fetch database rows by table", zap.Error(err))
	} else {
		for _, table := range databaseRowsByTableMetrics {
			for k, v := range table.stats {
				if f, ok := p.parseFloat(k, v); ok {
					labels := pdata.NewStringMap()
					labels.Insert(metadata.L.Database, table.database)
					labels.Insert(metadata.L.Table, table.table)
					labels.Insert(metadata.L.State, k)
					addToMetric(databaseRows, labels, f, now)
				}
			}
		}
	}

	// operations query
	operationsMetric, err := p.client.getOperations()
	if err != nil {
		p.logger.Error("Failed to fetch operations", zap.Error(err))
	} else {
		for k, v := range operationsMetric.stats {
			labels := pdata.NewStringMap()
			if i, ok := p.parseInt(k, v); ok {
				labels.Insert(metadata.L.Database, operationsMetric.database)
				labels.Insert(metadata.L.Table, operationsMetric.table)
				labels.Insert(metadata.L.Operation, k)
				addToIntMetric(operations, labels, i, now)
			}
		}
	}

	// operations by table
	operationsByTableMetrics, err := p.client.getOperationsByTable()
	if err != nil {
		p.logger.Error("Failed to fetch operations by table", zap.Error(err))
	} else {
		for _, table := range operationsByTableMetrics {
			for k, v := range table.stats {
				if i, ok := p.parseInt(k, v); ok {
					labels := pdata.NewStringMap()
					labels.Insert(metadata.L.Database, table.database)
					labels.Insert(metadata.L.Table, table.table)
					labels.Insert(metadata.L.Operation, k)
					addToIntMetric(operations, labels, i, now)
				}
			}
		}
	}

	// rollbacks
	rollbacksMetric, err := p.client.getRollbacks()
	if err != nil {
		p.logger.Error("Failed to fetch rollbacks", zap.Error(err))
	} else {
		for k, v := range rollbacksMetric.stats {
			labels := pdata.NewStringMap()
			if f, ok := p.parseFloat(k, v); ok {
				labels.Insert(metadata.L.Database, rollbacksMetric.database)
				addToMetric(rollbacks, labels, f, now)
			}
		}
	}

	return rms, nil
}

// parseFloat converts string to float64.
func (p *postgreSQLScraper) parseFloat(key, value string) (float64, bool) {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		p.logger.Info(
			"invalid value",
			zap.String("expectedType", "float"),
			zap.String("key", key),
			zap.String("value", value),
		)
		return 0, false
	}
	return f, true
}

// parseInt converts string to int64.
func (p *postgreSQLScraper) parseInt(key, value string) (int64, bool) {
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		p.logger.Info(
			"invalid value",
			zap.String("expectedType", "int"),
			zap.String("key", key),
			zap.String("value", value),
		)
		return 0, false
	}
	return i, true
}
