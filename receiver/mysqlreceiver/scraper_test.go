package mysqlreceiver

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/observiq/opentelemetry-components/receiver/mysqlreceiver/internal/metadata"
)

func TestScraperWithDatabase(t *testing.T) {
	mysqlMock := fakeClient{}
	sc := newMySQLScraper(zap.NewNop(), &Config{
		Username: "otel",
		Password: "otel",
		Database: "otel",
		Endpoint: "localhost:3306",
	})
	sc.client = &mysqlMock

	rms, err := sc.scrape(context.Background())
	require.Nil(t, err)

	require.Equal(t, 1, rms.Len())
	rm := rms.At(0)

	ilms := rm.InstrumentationLibraryMetrics()
	require.Equal(t, 1, ilms.Len())

	ilm := ilms.At(0)
	ms := ilm.Metrics()

	require.Equal(t, len(metadata.M.Names()), ms.Len())

	for i := 0; i < ms.Len(); i++ {
		m := ms.At(i)
		switch m.Name() {
		case metadata.M.MysqlBufferPoolPages.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 6, dps.Len())
			bufferPoolPagesMetrics := map[string]float64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolPages)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				bufferPoolPagesMetrics[label] = dp.DoubleVal()
			}
			require.Equal(t, 6, len(bufferPoolPagesMetrics))
			require.Equal(t, map[string]float64{
				"mysql.buffer_pool_pages :data database:otel":    981,
				"mysql.buffer_pool_pages :dirty database:otel":   0,
				"mysql.buffer_pool_pages :flushed database:otel": 168,
				"mysql.buffer_pool_pages :free database:otel":    7207,
				"mysql.buffer_pool_pages :misc database:otel":    4,
				"mysql.buffer_pool_pages :total database:otel":   8192,
			}, bufferPoolPagesMetrics)
		case metadata.M.MysqlBufferPoolOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 7, dps.Len())
			bufferPoolOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolOperations)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				bufferPoolOperationsMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 7, len(bufferPoolOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.buffer_pool_operations :read_ahead database:otel":         0,
				"mysql.buffer_pool_operations :read_ahead_evicted database:otel": 0,
				"mysql.buffer_pool_operations :read_ahead_rnd database:otel":     0,
				"mysql.buffer_pool_operations :read_requests database:otel":      14837,
				"mysql.buffer_pool_operations :reads database:otel":              838,
				"mysql.buffer_pool_operations :wait_free database:otel":          0,
				"mysql.buffer_pool_operations :write_requests database:otel":     1669,
			}, bufferPoolOperationsMetrics)
		case metadata.M.MysqlBufferPoolSize.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 3, dps.Len())
			bufferPoolSizeMetrics := map[string]float64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolSize)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				bufferPoolSizeMetrics[label] = dp.DoubleVal()
			}
			require.Equal(t, 3, len(bufferPoolSizeMetrics))
			require.Equal(t, map[string]float64{
				"mysql.buffer_pool_size :data database:otel":  16072704,
				"mysql.buffer_pool_size :dirty database:otel": 0,
				"mysql.buffer_pool_size :size database:otel":  134217728,
			}, bufferPoolSizeMetrics)
		case metadata.M.MysqlCommands.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 6, dps.Len())
			commandsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Command)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				commandsMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 6, len(commandsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.commands :close database:otel":          0,
				"mysql.commands :execute database:otel":        0,
				"mysql.commands :fetch database:otel":          0,
				"mysql.commands :prepare database:otel":        0,
				"mysql.commands :reset database:otel":          0,
				"mysql.commands :send_long_data database:otel": 0,
			}, commandsMetrics)
		case metadata.M.MysqlHandlers.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 18, dps.Len())
			handlersMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Handler)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				handlersMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 18, len(handlersMetrics))
			require.Equal(t, map[string]int64{
				"mysql.handlers :commit database:otel":             564,
				"mysql.handlers :delete database:otel":             0,
				"mysql.handlers :discover database:otel":           0,
				"mysql.handlers :lock database:otel":               7127,
				"mysql.handlers :mrr_init database:otel":           0,
				"mysql.handlers :prepare database:otel":            0,
				"mysql.handlers :read_first database:otel":         52,
				"mysql.handlers :read_key database:otel":           1680,
				"mysql.handlers :read_last database:otel":          0,
				"mysql.handlers :read_next database:otel":          3960,
				"mysql.handlers :read_prev database:otel":          0,
				"mysql.handlers :read_rnd database:otel":           0,
				"mysql.handlers :read_rnd_next database:otel":      505063,
				"mysql.handlers :rollback database:otel":           0,
				"mysql.handlers :savepoint database:otel":          0,
				"mysql.handlers :savepoint_rollback database:otel": 0,
				"mysql.handlers :update database:otel":             315,
				"mysql.handlers :write database:otel":              256734,
			}, handlersMetrics)
		case metadata.M.MysqlDoubleWrites.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			doubleWritesMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.DoubleWrites)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				doubleWritesMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 2, len(doubleWritesMetrics))
			require.Equal(t, map[string]int64{
				"mysql.double_writes :writes database:otel":  8,
				"mysql.double_writes :written database:otel": 27,
			}, doubleWritesMetrics)
		case metadata.M.MysqlLogOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			logOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.LogOperations)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				logOperationsMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 3, len(logOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.log_operations :requests database:otel": 646,
				"mysql.log_operations :waits database:otel":    0,
				"mysql.log_operations :writes database:otel":   14,
			}, logOperationsMetrics)
		case metadata.M.MysqlOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			operationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Operations)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				operationsMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 3, len(operationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.operations :fsyncs database:otel": 46,
				"mysql.operations :reads database:otel":  860,
				"mysql.operations :writes database:otel": 215,
			}, operationsMetrics)
		case metadata.M.MysqlPageOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			pageOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.PageOperations)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				pageOperationsMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 3, len(pageOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.page_operations :created database:otel": 144,
				"mysql.page_operations :read database:otel":    837,
				"mysql.page_operations :written database:otel": 168,
			}, pageOperationsMetrics)
		case metadata.M.MysqlRowLocks.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			rowLocksMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.RowLocks)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				rowLocksMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 2, len(rowLocksMetrics))
			require.Equal(t, map[string]int64{
				"mysql.row_locks :time database:otel":  0,
				"mysql.row_locks :waits database:otel": 0,
			}, rowLocksMetrics)
		case metadata.M.MysqlRowOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 4, dps.Len())
			rowOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.RowOperations)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				rowOperationsMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 4, len(rowOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.row_operations :deleted database:otel":  0,
				"mysql.row_operations :inserted database:otel": 0,
				"mysql.row_operations :read database:otel":     0,
				"mysql.row_operations :updated database:otel":  0,
			}, rowOperationsMetrics)
		case metadata.M.MysqlLocks.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			locksMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Locks)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				locksMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 2, len(locksMetrics))
			require.Equal(t, map[string]int64{
				"mysql.locks :immediate database:otel": 521,
				"mysql.locks :waited database:otel":    0,
			}, locksMetrics)
		case metadata.M.MysqlSorts.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 4, dps.Len())
			sortsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Sorts)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				sortsMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 4, len(sortsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.sorts :merge_passes database:otel": 0,
				"mysql.sorts :range database:otel":        0,
				"mysql.sorts :rows database:otel":         0,
				"mysql.sorts :scan database:otel":         0,
			}, sortsMetrics)
		case metadata.M.MysqlThreads.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 4, dps.Len())
			threadsMetrics := map[string]float64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Threads)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				threadsMetrics[label] = dp.DoubleVal()
			}
			require.Equal(t, 4, len(threadsMetrics))
			require.Equal(t, map[string]float64{
				"mysql.threads :cached database:otel":    0,
				"mysql.threads :connected database:otel": 1,
				"mysql.threads :created database:otel":   1,
				"mysql.threads :running database:otel":   2,
			}, threadsMetrics)
		}
	}
}

func TestScraperNoDatabase(t *testing.T) {
	mysqlMock := fakeClient{}
	sc := newMySQLScraper(zap.NewNop(), &Config{
		Username: "otel",
		Password: "otel",
		Endpoint: "localhost:3306",
	})
	sc.client = &mysqlMock

	rms, err := sc.scrape(context.Background())
	require.Nil(t, err)

	require.Equal(t, 1, rms.Len())
	rm := rms.At(0)

	ilms := rm.InstrumentationLibraryMetrics()
	require.Equal(t, 1, ilms.Len())

	ilm := ilms.At(0)
	ms := ilm.Metrics()

	require.Equal(t, len(metadata.M.Names()), ms.Len())

	for i := 0; i < ms.Len(); i++ {
		m := ms.At(i)
		switch m.Name() {
		case metadata.M.MysqlBufferPoolPages.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 6, dps.Len())
			bufferPoolPagesMetrics := map[string]float64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolPages)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				bufferPoolPagesMetrics[label] = dp.DoubleVal()
			}
			require.Equal(t, 6, len(bufferPoolPagesMetrics))
			require.Equal(t, map[string]float64{
				"mysql.buffer_pool_pages :data database:_global":    981,
				"mysql.buffer_pool_pages :dirty database:_global":   0,
				"mysql.buffer_pool_pages :flushed database:_global": 168,
				"mysql.buffer_pool_pages :free database:_global":    7207,
				"mysql.buffer_pool_pages :misc database:_global":    4,
				"mysql.buffer_pool_pages :total database:_global":   8192,
			}, bufferPoolPagesMetrics)
		case metadata.M.MysqlBufferPoolOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 7, dps.Len())
			bufferPoolOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolOperations)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				bufferPoolOperationsMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 7, len(bufferPoolOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.buffer_pool_operations :read_ahead database:_global":         0,
				"mysql.buffer_pool_operations :read_ahead_evicted database:_global": 0,
				"mysql.buffer_pool_operations :read_ahead_rnd database:_global":     0,
				"mysql.buffer_pool_operations :read_requests database:_global":      14837,
				"mysql.buffer_pool_operations :reads database:_global":              838,
				"mysql.buffer_pool_operations :wait_free database:_global":          0,
				"mysql.buffer_pool_operations :write_requests database:_global":     1669,
			}, bufferPoolOperationsMetrics)
		case metadata.M.MysqlBufferPoolSize.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 3, dps.Len())
			bufferPoolSizeMetrics := map[string]float64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolSize)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				bufferPoolSizeMetrics[label] = dp.DoubleVal()
			}
			require.Equal(t, 3, len(bufferPoolSizeMetrics))
			require.Equal(t, map[string]float64{
				"mysql.buffer_pool_size :data database:_global":  16072704,
				"mysql.buffer_pool_size :dirty database:_global": 0,
				"mysql.buffer_pool_size :size database:_global":  134217728,
			}, bufferPoolSizeMetrics)
		case metadata.M.MysqlCommands.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 6, dps.Len())
			commandsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Command)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				commandsMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 6, len(commandsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.commands :close database:_global":          0,
				"mysql.commands :execute database:_global":        0,
				"mysql.commands :fetch database:_global":          0,
				"mysql.commands :prepare database:_global":        0,
				"mysql.commands :reset database:_global":          0,
				"mysql.commands :send_long_data database:_global": 0,
			}, commandsMetrics)
		case metadata.M.MysqlHandlers.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 18, dps.Len())
			handlersMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Handler)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				handlersMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 18, len(handlersMetrics))
			require.Equal(t, map[string]int64{
				"mysql.handlers :commit database:_global":             564,
				"mysql.handlers :delete database:_global":             0,
				"mysql.handlers :discover database:_global":           0,
				"mysql.handlers :lock database:_global":               7127,
				"mysql.handlers :mrr_init database:_global":           0,
				"mysql.handlers :prepare database:_global":            0,
				"mysql.handlers :read_first database:_global":         52,
				"mysql.handlers :read_key database:_global":           1680,
				"mysql.handlers :read_last database:_global":          0,
				"mysql.handlers :read_next database:_global":          3960,
				"mysql.handlers :read_prev database:_global":          0,
				"mysql.handlers :read_rnd database:_global":           0,
				"mysql.handlers :read_rnd_next database:_global":      505063,
				"mysql.handlers :rollback database:_global":           0,
				"mysql.handlers :savepoint database:_global":          0,
				"mysql.handlers :savepoint_rollback database:_global": 0,
				"mysql.handlers :update database:_global":             315,
				"mysql.handlers :write database:_global":              256734,
			}, handlersMetrics)
		case metadata.M.MysqlDoubleWrites.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			doubleWritesMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.DoubleWrites)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				doubleWritesMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 2, len(doubleWritesMetrics))
			require.Equal(t, map[string]int64{
				"mysql.double_writes :writes database:_global":  8,
				"mysql.double_writes :written database:_global": 27,
			}, doubleWritesMetrics)
		case metadata.M.MysqlLogOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			logOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.LogOperations)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				logOperationsMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 3, len(logOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.log_operations :requests database:_global": 646,
				"mysql.log_operations :waits database:_global":    0,
				"mysql.log_operations :writes database:_global":   14,
			}, logOperationsMetrics)
		case metadata.M.MysqlOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			operationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Operations)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				operationsMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 3, len(operationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.operations :fsyncs database:_global": 46,
				"mysql.operations :reads database:_global":  860,
				"mysql.operations :writes database:_global": 215,
			}, operationsMetrics)
		case metadata.M.MysqlPageOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			pageOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.PageOperations)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				pageOperationsMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 3, len(pageOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.page_operations :created database:_global": 144,
				"mysql.page_operations :read database:_global":    837,
				"mysql.page_operations :written database:_global": 168,
			}, pageOperationsMetrics)
		case metadata.M.MysqlRowLocks.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			rowLocksMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.RowLocks)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				rowLocksMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 2, len(rowLocksMetrics))
			require.Equal(t, map[string]int64{
				"mysql.row_locks :time database:_global":  0,
				"mysql.row_locks :waits database:_global": 0,
			}, rowLocksMetrics)
		case metadata.M.MysqlRowOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 4, dps.Len())
			rowOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.RowOperations)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				rowOperationsMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 4, len(rowOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.row_operations :deleted database:_global":  0,
				"mysql.row_operations :inserted database:_global": 0,
				"mysql.row_operations :read database:_global":     0,
				"mysql.row_operations :updated database:_global":  0,
			}, rowOperationsMetrics)
		case metadata.M.MysqlLocks.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			locksMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Locks)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				locksMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 2, len(locksMetrics))
			require.Equal(t, map[string]int64{
				"mysql.locks :immediate database:_global": 521,
				"mysql.locks :waited database:_global":    0,
			}, locksMetrics)
		case metadata.M.MysqlSorts.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 4, dps.Len())
			sortsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Sorts)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				sortsMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 4, len(sortsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.sorts :merge_passes database:_global": 0,
				"mysql.sorts :range database:_global":        0,
				"mysql.sorts :rows database:_global":         0,
				"mysql.sorts :scan database:_global":         0,
			}, sortsMetrics)
		case metadata.M.MysqlThreads.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 4, dps.Len())
			threadsMetrics := map[string]float64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Threads)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				threadsMetrics[label] = dp.DoubleVal()
			}
			require.Equal(t, 4, len(threadsMetrics))
			require.Equal(t, map[string]float64{
				"mysql.threads :cached database:_global":    0,
				"mysql.threads :connected database:_global": 1,
				"mysql.threads :created database:_global":   1,
				"mysql.threads :running database:_global":   2,
			}, threadsMetrics)
		}
	}
}
