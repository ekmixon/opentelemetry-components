package mysqlreceiver

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"

	"github.com/observiq/opentelemetry-components/receiver/mysqlreceiver/internal/metadata"
)

type mySQLScraper struct {
	client   client
	stopOnce sync.Once

	logger *zap.Logger
	config *Config
}

func newMySQLScraper(
	logger *zap.Logger,
	config *Config,
) *mySQLScraper {
	return &mySQLScraper{
		logger: logger,
		config: config,
	}
}

// start starts the scraper by initializing the db client connection.
func (m *mySQLScraper) start(_ context.Context, host component.Host) error {
	client, err := newMySQLClient(mySQLConfig{
		username: m.config.Username,
		password: m.config.Password,
		endpoint: m.config.Endpoint,
	})
	if err != nil {
		return err
	}
	m.client = client

	return nil
}

// shutdown closes the db connection
func (m *mySQLScraper) shutdown(context.Context) error {
	var err error
	m.stopOnce.Do(func() {
		err = m.client.Close()
	})
	return err
}

// initMetric initializes a metric with a metadata label.
func initMetric(ms pdata.MetricSlice, mi metadata.MetricIntf) pdata.Metric {
	m := ms.AppendEmpty()
	mi.Init(m)
	return m
}

// addToDoubleMetric adds and labels a double gauge datapoint to a metricslice.
func addToDoubleMetric(metric pdata.NumberDataPointSlice, labels pdata.StringMap, value float64, ts pdata.Timestamp) {
	dataPoint := metric.AppendEmpty()
	dataPoint.SetTimestamp(ts)
	dataPoint.SetDoubleVal(value)
	if labels.Len() > 0 {
		labels.CopyTo(dataPoint.LabelsMap())
	}
}

// addToIntMetric adds and labels a int sum datapoint to metricslice.
func addToIntMetric(metric pdata.NumberDataPointSlice, labels pdata.StringMap, value int64, ts pdata.Timestamp) {
	dataPoint := metric.AppendEmpty()
	dataPoint.SetTimestamp(ts)
	dataPoint.SetIntVal(value)
	if labels.Len() > 0 {
		labels.CopyTo(dataPoint.LabelsMap())
	}
}

// scrape scrapes the mysql db metric stats, transforms them and labels them into a metric slices.
func (m *mySQLScraper) scrape(context.Context) (pdata.ResourceMetricsSlice, error) {

	if m.client == nil {
		return pdata.ResourceMetricsSlice{}, errors.New("failed to connect to http client")
	}

	// metric initialization
	rms := pdata.NewResourceMetricsSlice()
	ilm := rms.AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty()
	ilm.InstrumentationLibrary().SetName("otel/mysql")
	now := pdata.TimestampFromTime(time.Now())

	bufferPoolPages := initMetric(ilm.Metrics(), metadata.M.MysqlBufferPoolPages).Gauge().DataPoints()
	bufferPoolOperations := initMetric(ilm.Metrics(), metadata.M.MysqlBufferPoolOperations).Sum().DataPoints()
	bufferPoolSize := initMetric(ilm.Metrics(), metadata.M.MysqlBufferPoolSize).Gauge().DataPoints()
	commands := initMetric(ilm.Metrics(), metadata.M.MysqlCommands).Sum().DataPoints()
	handlers := initMetric(ilm.Metrics(), metadata.M.MysqlHandlers).Sum().DataPoints()
	doubleWrites := initMetric(ilm.Metrics(), metadata.M.MysqlDoubleWrites).Sum().DataPoints()
	logOperations := initMetric(ilm.Metrics(), metadata.M.MysqlLogOperations).Sum().DataPoints()
	operations := initMetric(ilm.Metrics(), metadata.M.MysqlOperations).Sum().DataPoints()
	pageOperations := initMetric(ilm.Metrics(), metadata.M.MysqlPageOperations).Sum().DataPoints()
	rowLocks := initMetric(ilm.Metrics(), metadata.M.MysqlRowLocks).Sum().DataPoints()
	rowOperations := initMetric(ilm.Metrics(), metadata.M.MysqlRowOperations).Sum().DataPoints()
	locks := initMetric(ilm.Metrics(), metadata.M.MysqlLocks).Sum().DataPoints()
	sorts := initMetric(ilm.Metrics(), metadata.M.MysqlSorts).Sum().DataPoints()
	threads := initMetric(ilm.Metrics(), metadata.M.MysqlThreads).Gauge().DataPoints()

	// collect innodb metrics.
	innodbStats, err := m.client.getInnodbStats()
	for k, v := range innodbStats {
		labels := pdata.NewStringMap()
		switch k {
		case "buffer_pool_size":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.BufferPoolSize, "size")
				addToDoubleMetric(bufferPoolSize, labels, f, now)
			}
		}
	}
	if err != nil {
		m.logger.Error("Failed to fetch InnoDB stats", zap.Error(err))
		return pdata.ResourceMetricsSlice{}, err
	}

	// collect global status metrics.
	globalStats, err := m.client.getGlobalStats()
	if err != nil {
		m.logger.Error("Failed to fetch global stats", zap.Error(err))
		return pdata.ResourceMetricsSlice{}, err
	}

	for k, v := range globalStats {
		labels := pdata.NewStringMap()
		switch k {

		// buffer_pool_pages
		case "Innodb_buffer_pool_pages_data":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.BufferPoolPages, "data")
				addToDoubleMetric(bufferPoolPages, labels, f, now)
			}
		case "Innodb_buffer_pool_pages_dirty":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.BufferPoolPages, "dirty")
				addToDoubleMetric(bufferPoolPages, labels, f, now)
			}
		case "Innodb_buffer_pool_pages_flushed":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.BufferPoolPages, "flushed")
				addToDoubleMetric(bufferPoolPages, labels, f, now)
			}
		case "Innodb_buffer_pool_pages_free":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.BufferPoolPages, "free")
				addToDoubleMetric(bufferPoolPages, labels, f, now)
			}
		case "Innodb_buffer_pool_pages_misc":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.BufferPoolPages, "misc")
				addToDoubleMetric(bufferPoolPages, labels, f, now)
			}
		case "Innodb_buffer_pool_pages_total":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.BufferPoolPages, "total")
				addToDoubleMetric(bufferPoolPages, labels, f, now)
			}

		// buffer_pool_operations
		case "Innodb_buffer_pool_read_ahead_rnd":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.BufferPoolOperations, "read_ahead_rnd")
				addToIntMetric(bufferPoolOperations, labels, i, now)
			}
		case "Innodb_buffer_pool_read_ahead":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.BufferPoolOperations, "read_ahead")
				addToIntMetric(bufferPoolOperations, labels, i, now)
			}
		case "Innodb_buffer_pool_read_ahead_evicted":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.BufferPoolOperations, "read_ahead_evicted")
				addToIntMetric(bufferPoolOperations, labels, i, now)
			}
		case "Innodb_buffer_pool_read_requests":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.BufferPoolOperations, "read_requests")
				addToIntMetric(bufferPoolOperations, labels, i, now)
			}
		case "Innodb_buffer_pool_reads":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.BufferPoolOperations, "reads")
				addToIntMetric(bufferPoolOperations, labels, i, now)
			}
		case "Innodb_buffer_pool_wait_free":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.BufferPoolOperations, "wait_free")
				addToIntMetric(bufferPoolOperations, labels, i, now)
			}
		case "Innodb_buffer_pool_write_requests":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.BufferPoolOperations, "write_requests")
				addToIntMetric(bufferPoolOperations, labels, i, now)
			}

		// buffer_pool_size
		case "Innodb_buffer_pool_bytes_data":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.BufferPoolSize, "data")
				addToDoubleMetric(bufferPoolSize, labels, f, now)
			}
		case "Innodb_buffer_pool_bytes_dirty":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.BufferPoolSize, "dirty")
				addToDoubleMetric(bufferPoolSize, labels, f, now)
			}

		// commands
		case "Com_stmt_execute":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Command, "execute")
				addToIntMetric(commands, labels, i, now)
			}
		case "Com_stmt_close":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Command, "close")
				addToIntMetric(commands, labels, i, now)
			}
		case "Com_stmt_fetch":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Command, "fetch")
				addToIntMetric(commands, labels, i, now)
			}
		case "Com_stmt_prepare":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Command, "prepare")
				addToIntMetric(commands, labels, i, now)
			}
		case "Com_stmt_reset":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Command, "reset")
				addToIntMetric(commands, labels, i, now)
			}
		case "Com_stmt_send_long_data":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Command, "send_long_data")
				addToIntMetric(commands, labels, i, now)
			}

		// handlers
		case "Handler_commit":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Handler, "commit")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_delete":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Handler, "delete")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_discover":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Handler, "discover")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_external_lock":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Handler, "lock")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_mrr_init":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Handler, "mrr_init")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_prepare":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Handler, "prepare")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_read_first":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Handler, "read_first")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_read_key":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Handler, "read_key")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_read_last":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Handler, "read_last")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_read_next":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Handler, "read_next")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_read_prev":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Handler, "read_prev")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_read_rnd":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Handler, "read_rnd")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_read_rnd_next":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Handler, "read_rnd_next")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_rollback":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Handler, "rollback")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_savepoint":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Handler, "savepoint")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_savepoint_rollback":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Handler, "savepoint_rollback")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_update":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Handler, "update")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_write":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Handler, "write")
				addToIntMetric(handlers, labels, i, now)
			}

		// double_writes
		case "Innodb_dblwr_pages_written":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.DoubleWrites, "written")
				addToIntMetric(doubleWrites, labels, i, now)
			}
		case "Innodb_dblwr_writes":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.DoubleWrites, "writes")
				addToIntMetric(doubleWrites, labels, i, now)
			}

		// log_operations
		case "Innodb_log_waits":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.LogOperations, "waits")
				addToIntMetric(logOperations, labels, i, now)
			}
		case "Innodb_log_write_requests":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.LogOperations, "requests")
				addToIntMetric(logOperations, labels, i, now)
			}
		case "Innodb_log_writes":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.LogOperations, "writes")
				addToIntMetric(logOperations, labels, i, now)
			}

		// operations
		case "Innodb_data_fsyncs":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Operations, "fsyncs")
				addToIntMetric(operations, labels, i, now)
			}
		case "Innodb_data_reads":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Operations, "reads")
				addToIntMetric(operations, labels, i, now)
			}
		case "Innodb_data_writes":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Operations, "writes")
				addToIntMetric(operations, labels, i, now)
			}

		// page_operations
		case "Innodb_pages_created":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.PageOperations, "created")
				addToIntMetric(pageOperations, labels, i, now)
			}
		case "Innodb_pages_read":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.PageOperations, "read")
				addToIntMetric(pageOperations, labels, i, now)
			}
		case "Innodb_pages_written":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.PageOperations, "written")
				addToIntMetric(pageOperations, labels, i, now)
			}

		// row_locks
		case "Innodb_row_lock_waits":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.RowLocks, "waits")
				addToIntMetric(rowLocks, labels, i, now)
			}
		case "Innodb_row_lock_time":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.RowLocks, "time")
				addToIntMetric(rowLocks, labels, i, now)
			}

		// row_operations
		case "Innodb_rows_deleted":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.RowOperations, "deleted")
				addToIntMetric(rowOperations, labels, i, now)
			}
		case "Innodb_rows_inserted":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.RowOperations, "inserted")
				addToIntMetric(rowOperations, labels, i, now)
			}
		case "Innodb_rows_read":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.RowOperations, "read")
				addToIntMetric(rowOperations, labels, i, now)
			}
		case "Innodb_rows_updated":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.RowOperations, "updated")
				addToIntMetric(rowOperations, labels, i, now)
			}

		// locks
		case "Table_locks_immediate":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Locks, "immediate")
				addToIntMetric(locks, labels, i, now)
			}
		case "Table_locks_waited":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Locks, "waited")
				addToIntMetric(locks, labels, i, now)
			}

		// sorts
		case "Sort_merge_passes":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Sorts, "merge_passes")
				addToIntMetric(sorts, labels, i, now)
			}
		case "Sort_range":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Sorts, "range")
				addToIntMetric(sorts, labels, i, now)
			}
		case "Sort_rows":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Sorts, "rows")
				addToIntMetric(sorts, labels, i, now)
			}
		case "Sort_scan":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.Sorts, "scan")
				addToIntMetric(sorts, labels, i, now)
			}

		// threads
		case "Threads_cached":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.Threads, "cached")
				addToDoubleMetric(threads, labels, f, now)
			}
		case "Threads_connected":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.Threads, "connected")
				addToDoubleMetric(threads, labels, f, now)
			}
		case "Threads_created":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.Threads, "created")
				addToDoubleMetric(threads, labels, f, now)
			}
		case "Threads_running":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.Threads, "running")
				addToDoubleMetric(threads, labels, f, now)
			}
		}
	}
	return rms, nil
}

// parseFloat converts string to float64.
func (m *mySQLScraper) parseFloat(key, value string) (float64, bool) {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		m.logInvalid("float", key, value)
		return 0, false
	}
	return f, true
}

// parseInt converts string to int64.
func (m *mySQLScraper) parseInt(key, value string) (int64, bool) {
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		m.logInvalid("int", key, value)
		return 0, false
	}
	return i, true
}

func (m *mySQLScraper) logInvalid(expectedType, key, value string) {
	m.logger.Info(
		"invalid value",
		zap.String("expectedType", expectedType),
		zap.String("key", key),
		zap.String("value", value),
	)
}
