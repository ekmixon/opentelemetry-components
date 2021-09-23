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
		username: p.config.Username,
		password: p.config.Password,
		database: p.config.Database,
		endpoint: p.config.Endpoint,
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

// addToIntMetric adds and labels a int sum datapoint to metricslice.
func addToIntMetric(metric pdata.NumberDataPointSlice, labels pdata.AttributeMap, value int64, ts pdata.Timestamp) {
	dataPoint := metric.AppendEmpty()
	dataPoint.SetTimestamp(ts)
	dataPoint.SetIntVal(value)
	if labels.Len() > 0 {
		labels.CopyTo(dataPoint.Attributes())
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
	now := pdata.NewTimestampFromTime(time.Now())

	blocks_read := initMetric(ilm.Metrics(), metadata.M.PostgresqlBlocksRead).Sum().DataPoints()
	commits := initMetric(ilm.Metrics(), metadata.M.PostgresqlCommits).Sum().DataPoints()
	databaseSize := initMetric(ilm.Metrics(), metadata.M.PostgresqlDbSize).Gauge().DataPoints()
	backends := initMetric(ilm.Metrics(), metadata.M.PostgresqlBackends).Gauge().DataPoints()
	databaseRows := initMetric(ilm.Metrics(), metadata.M.PostgresqlRows).Gauge().DataPoints()
	operations := initMetric(ilm.Metrics(), metadata.M.PostgresqlOperations).Sum().DataPoints()
	rollbacks := initMetric(ilm.Metrics(), metadata.M.PostgresqlRollbacks).Gauge().DataPoints()

	// blocks read query
	blocksReadMetric, err := p.client.getBlocksRead()
	if err != nil {
		p.logger.Error("Failed to fetch blocks read", zap.Error(err))
	} else {
		for k, v := range blocksReadMetric.stats {
			labels := pdata.NewAttributeMap()
			if i, ok := p.parseInt(k, v); ok {
				labels.Insert(metadata.L.Database, pdata.NewAttributeValueString(blocksReadMetric.database))
				labels.Insert(metadata.L.Table, pdata.NewAttributeValueString(blocksReadMetric.table))
				labels.Insert(metadata.L.Source, pdata.NewAttributeValueString(k))
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
					labels := pdata.NewAttributeMap()
					labels.Insert(metadata.L.Database, pdata.NewAttributeValueString(table.database))
					labels.Insert(metadata.L.Table, pdata.NewAttributeValueString(table.table))
					labels.Insert(metadata.L.Source, pdata.NewAttributeValueString(k))
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
			labels := pdata.NewAttributeMap()
			if i, ok := p.parseInt(k, v); ok {
				labels.Insert(metadata.L.Database, pdata.NewAttributeValueString(commitsMetric.database))
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
			labels := pdata.NewAttributeMap()
			if f, ok := p.parseInt(k, v); ok {
				labels.Insert(metadata.L.Database, pdata.NewAttributeValueString(databaseSizeMetric.database))
				addToIntMetric(databaseSize, labels, f, now)
			}
		}
	}

	// backends
	backendsMetric, err := p.client.getBackends()
	if err != nil {
		p.logger.Error("Failed to fetch backends", zap.Error(err))
	} else {
		for k, v := range backendsMetric.stats {
			labels := pdata.NewAttributeMap()
			if f, ok := p.parseInt(k, v); ok {
				labels.Insert(metadata.L.Database, pdata.NewAttributeValueString(backendsMetric.database))
				addToIntMetric(backends, labels, f, now)
			}
		}
	}

	// database rows query
	databaseRowsMetric, err := p.client.getDatabaseRows()
	if err != nil {
		p.logger.Error("Failed to fetch database rows", zap.Error(err))
	} else {
		for k, v := range databaseRowsMetric.stats {
			labels := pdata.NewAttributeMap()
			if f, ok := p.parseInt(k, v); ok {
				labels.Insert(metadata.L.Database, pdata.NewAttributeValueString(databaseRowsMetric.database))
				labels.Insert(metadata.L.Table, pdata.NewAttributeValueString(databaseRowsMetric.table))
				labels.Insert(metadata.L.State, pdata.NewAttributeValueString(k))
				addToIntMetric(databaseRows, labels, f, now)
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
				if f, ok := p.parseInt(k, v); ok {
					labels := pdata.NewAttributeMap()
					labels.Insert(metadata.L.Database, pdata.NewAttributeValueString(table.database))
					labels.Insert(metadata.L.Table, pdata.NewAttributeValueString(table.table))
					labels.Insert(metadata.L.State, pdata.NewAttributeValueString(k))
					addToIntMetric(databaseRows, labels, f, now)
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
			labels := pdata.NewAttributeMap()
			if i, ok := p.parseInt(k, v); ok {
				labels.Insert(metadata.L.Database, pdata.NewAttributeValueString(operationsMetric.database))
				labels.Insert(metadata.L.Table, pdata.NewAttributeValueString(operationsMetric.table))
				labels.Insert(metadata.L.Operation, pdata.NewAttributeValueString(k))
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
					labels := pdata.NewAttributeMap()
					labels.Insert(metadata.L.Database, pdata.NewAttributeValueString(table.database))
					labels.Insert(metadata.L.Table, pdata.NewAttributeValueString(table.table))
					labels.Insert(metadata.L.Operation, pdata.NewAttributeValueString(k))
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
			labels := pdata.NewAttributeMap()
			if f, ok := p.parseInt(k, v); ok {
				labels.Insert(metadata.L.Database, pdata.NewAttributeValueString(rollbacksMetric.database))
				addToIntMetric(rollbacks, labels, f, now)
			}
		}
	}

	return rms, nil
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
