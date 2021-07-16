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
		database: m.config.Database,
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
	bufferPoolOperations := initMetric(ilm.Metrics(), metadata.M.MysqlBufferPoolOperations).IntSum().DataPoints()
	bufferPoolSize := initMetric(ilm.Metrics(), metadata.M.MysqlBufferPoolSize).Gauge().DataPoints()
	commands := initMetric(ilm.Metrics(), metadata.M.MysqlCommands).IntSum().DataPoints()
	handlers := initMetric(ilm.Metrics(), metadata.M.MysqlHandlers).IntSum().DataPoints()
	doubleWrites := initMetric(ilm.Metrics(), metadata.M.MysqlDoubleWrites).IntSum().DataPoints()
	logOperations := initMetric(ilm.Metrics(), metadata.M.MysqlLogOperations).IntSum().DataPoints()
	operations := initMetric(ilm.Metrics(), metadata.M.MysqlOperations).IntSum().DataPoints()
	pageOperations := initMetric(ilm.Metrics(), metadata.M.MysqlPageOperations).IntSum().DataPoints()
	rowLocks := initMetric(ilm.Metrics(), metadata.M.MysqlRowLocks).IntSum().DataPoints()
	rowOperations := initMetric(ilm.Metrics(), metadata.M.MysqlRowOperations).IntSum().DataPoints()
	locks := initMetric(ilm.Metrics(), metadata.M.MysqlLocks).IntSum().DataPoints()
	sorts := initMetric(ilm.Metrics(), metadata.M.MysqlSorts).IntSum().DataPoints()
	threads := initMetric(ilm.Metrics(), metadata.M.MysqlThreads).Gauge().DataPoints()

	// collect innodb metrics.
	innodbStats, err := m.client.getInnodbStats()
	for k, v := range innodbStats {
		labels := pdata.NewStringMap()
		switch k {
		case "buffer_pool_size":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.BufferPoolSizeState, "size")
				addToMetric(bufferPoolSize, labels, f, now)
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
				labels.Insert(metadata.L.BufferPoolPagesState, "data")
				addToMetric(bufferPoolPages, labels, f, now)
			}
		case "Innodb_buffer_pool_pages_dirty":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.BufferPoolPagesState, "dirty")
				addToMetric(bufferPoolPages, labels, f, now)
			}
		case "Innodb_buffer_pool_pages_flushed":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.BufferPoolPagesState, "flushed")
				addToMetric(bufferPoolPages, labels, f, now)
			}
		case "Innodb_buffer_pool_pages_free":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.BufferPoolPagesState, "free")
				addToMetric(bufferPoolPages, labels, f, now)
			}
		case "Innodb_buffer_pool_pages_misc":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.BufferPoolPagesState, "misc")
				addToMetric(bufferPoolPages, labels, f, now)
			}
		case "Innodb_buffer_pool_pages_total":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.BufferPoolPagesState, "total")
				addToMetric(bufferPoolPages, labels, f, now)
			}

		// buffer_pool_operations
		case "Innodb_buffer_pool_read_ahead_rnd":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.BufferPoolOperationsState, "read_ahead_rnd")
				addToIntMetric(bufferPoolOperations, labels, i, now)
			}
		case "Innodb_buffer_pool_read_ahead":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.BufferPoolOperationsState, "read_ahead")
				addToIntMetric(bufferPoolOperations, labels, i, now)
			}
		case "Innodb_buffer_pool_read_ahead_evicted":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.BufferPoolOperationsState, "read_ahead_evicted")
				addToIntMetric(bufferPoolOperations, labels, i, now)
			}
		case "Innodb_buffer_pool_read_requests":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.BufferPoolOperationsState, "read_requests")
				addToIntMetric(bufferPoolOperations, labels, i, now)
			}
		case "Innodb_buffer_pool_reads":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.BufferPoolOperationsState, "reads")
				addToIntMetric(bufferPoolOperations, labels, i, now)
			}
		case "Innodb_buffer_pool_wait_free":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.BufferPoolOperationsState, "wait_free")
				addToIntMetric(bufferPoolOperations, labels, i, now)
			}
		case "Innodb_buffer_pool_write_requests":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.BufferPoolOperationsState, "write_requests")
				addToIntMetric(bufferPoolOperations, labels, i, now)
			}

		// buffer_pool_size
		case "Innodb_buffer_pool_bytes_data":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.BufferPoolSizeState, "data")
				addToMetric(bufferPoolSize, labels, f, now)
			}
		case "Innodb_buffer_pool_bytes_dirty":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.BufferPoolSizeState, "dirty")
				addToMetric(bufferPoolSize, labels, f, now)
			}

		// commands
		case "Com_stmt_execute":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.CommandState, "execute")
				addToIntMetric(commands, labels, i, now)
			}
		case "Com_stmt_close":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.CommandState, "close")
				addToIntMetric(commands, labels, i, now)
			}
		case "Com_stmt_fetch":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.CommandState, "fetch")
				addToIntMetric(commands, labels, i, now)
			}
		case "Com_stmt_prepare":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.CommandState, "prepare")
				addToIntMetric(commands, labels, i, now)
			}
		case "Com_stmt_reset":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.CommandState, "reset")
				addToIntMetric(commands, labels, i, now)
			}
		case "Com_stmt_send_long_data":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.CommandState, "send_long_data")
				addToIntMetric(commands, labels, i, now)
			}

		// handlers
		case "Handler_commit":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.HandlerState, "commit")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_delete":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.HandlerState, "delete")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_discover":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.HandlerState, "discover")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_external_lock":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.HandlerState, "lock")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_mrr_init":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.HandlerState, "mrr_init")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_prepare":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.HandlerState, "prepare")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_read_first":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.HandlerState, "read_first")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_read_key":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.HandlerState, "read_key")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_read_last":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.HandlerState, "read_last")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_read_next":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.HandlerState, "read_next")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_read_prev":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.HandlerState, "read_prev")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_read_rnd":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.HandlerState, "read_rnd")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_read_rnd_next":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.HandlerState, "read_rnd_next")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_rollback":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.HandlerState, "rollback")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_savepoint":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.HandlerState, "savepoint")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_savepoint_rollback":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.HandlerState, "savepoint_rollback")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_update":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.HandlerState, "update")
				addToIntMetric(handlers, labels, i, now)
			}
		case "Handler_write":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.HandlerState, "write")
				addToIntMetric(handlers, labels, i, now)
			}

		// double_writes
		case "Innodb_dblwr_pages_written":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.DoubleWritesState, "written")
				addToIntMetric(doubleWrites, labels, i, now)
			}
		case "Innodb_dblwr_writes":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.DoubleWritesState, "writes")
				addToIntMetric(doubleWrites, labels, i, now)
			}

		// log_operations
		case "Innodb_log_waits":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.LogOperationsState, "waits")
				addToIntMetric(logOperations, labels, i, now)
			}
		case "Innodb_log_write_requests":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.LogOperationsState, "requests")
				addToIntMetric(logOperations, labels, i, now)
			}
		case "Innodb_log_writes":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.LogOperationsState, "writes")
				addToIntMetric(logOperations, labels, i, now)
			}

		// operations
		case "Innodb_data_fsyncs":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.OperationsState, "fsyncs")
				addToIntMetric(operations, labels, i, now)
			}
		case "Innodb_data_reads":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.OperationsState, "reads")
				addToIntMetric(operations, labels, i, now)
			}
		case "Innodb_data_writes":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.OperationsState, "writes")
				addToIntMetric(operations, labels, i, now)
			}

		// page_operations
		case "Innodb_pages_created":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.PageOperationsState, "created")
				addToIntMetric(pageOperations, labels, i, now)
			}
		case "Innodb_pages_read":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.PageOperationsState, "read")
				addToIntMetric(pageOperations, labels, i, now)
			}
		case "Innodb_pages_written":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.PageOperationsState, "written")
				addToIntMetric(pageOperations, labels, i, now)
			}

		// row_locks
		case "Innodb_row_lock_waits":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.RowLocksState, "waits")
				addToIntMetric(rowLocks, labels, i, now)
			}
		case "Innodb_row_lock_time":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.RowLocksState, "time")
				addToIntMetric(rowLocks, labels, i, now)
			}

		// row_operations
		case "Innodb_rows_deleted":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.RowOperationsState, "deleted")
				addToIntMetric(rowOperations, labels, i, now)
			}
		case "Innodb_rows_inserted":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.RowOperationsState, "inserted")
				addToIntMetric(rowOperations, labels, i, now)
			}
		case "Innodb_rows_read":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.RowOperationsState, "read")
				addToIntMetric(rowOperations, labels, i, now)
			}
		case "Innodb_rows_updated":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.RowOperationsState, "updated")
				addToIntMetric(rowOperations, labels, i, now)
			}

		// locks
		case "Table_locks_immediate":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.LocksState, "immediate")
				addToIntMetric(locks, labels, i, now)
			}
		case "Table_locks_waited":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.LocksState, "waited")
				addToIntMetric(locks, labels, i, now)
			}

		// sorts
		case "Sort_merge_passes":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.SortsState, "merge_passes")
				addToIntMetric(sorts, labels, i, now)
			}
		case "Sort_range":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.SortsState, "range")
				addToIntMetric(sorts, labels, i, now)
			}
		case "Sort_rows":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.SortsState, "rows")
				addToIntMetric(sorts, labels, i, now)
			}
		case "Sort_scan":
			if i, ok := m.parseInt(k, v); ok {
				labels.Insert(metadata.L.SortsState, "scan")
				addToIntMetric(sorts, labels, i, now)
			}

		// threads
		case "Threads_cached":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.ThreadsState, "cached")
				addToMetric(threads, labels, f, now)
			}
		case "Threads_connected":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.ThreadsState, "connected")
				addToMetric(threads, labels, f, now)
			}
		case "Threads_created":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.ThreadsState, "created")
				addToMetric(threads, labels, f, now)
			}
		case "Threads_running":
			if f, ok := m.parseFloat(k, v); ok {
				labels.Insert(metadata.L.ThreadsState, "running")
				addToMetric(threads, labels, f, now)
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
