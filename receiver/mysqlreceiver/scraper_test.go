package mysqlreceiver

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/observiq/opentelemetry-components/receiver/mysqlreceiver/internal/metadata"
)

func TestScrape(t *testing.T) {
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
				label := fmt.Sprintf("%s :%s", m.Name(), value_label)
				bufferPoolPagesMetrics[label] = dp.DoubleVal()
			}
			require.Equal(t, 6, len(bufferPoolPagesMetrics))
			require.Equal(t, map[string]float64{
				"mysql.buffer_pool_pages :data":    981,
				"mysql.buffer_pool_pages :dirty":   0,
				"mysql.buffer_pool_pages :flushed": 168,
				"mysql.buffer_pool_pages :free":    7207,
				"mysql.buffer_pool_pages :misc":    4,
				"mysql.buffer_pool_pages :total":   8192,
			}, bufferPoolPagesMetrics)
		case metadata.M.MysqlBufferPoolOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 7, dps.Len())
			bufferPoolOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolOperations)
				label := fmt.Sprintf("%s :%s", m.Name(), value_label)
				bufferPoolOperationsMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 7, len(bufferPoolOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.buffer_pool_operations :read_ahead":         0,
				"mysql.buffer_pool_operations :read_ahead_evicted": 0,
				"mysql.buffer_pool_operations :read_ahead_rnd":     0,
				"mysql.buffer_pool_operations :read_requests":      14837,
				"mysql.buffer_pool_operations :reads":              838,
				"mysql.buffer_pool_operations :wait_free":          0,
				"mysql.buffer_pool_operations :write_requests":     1669,
			}, bufferPoolOperationsMetrics)
		case metadata.M.MysqlBufferPoolSize.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 3, dps.Len())
			bufferPoolSizeMetrics := map[string]float64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolSize)
				label := fmt.Sprintf("%s :%s", m.Name(), value_label)
				bufferPoolSizeMetrics[label] = dp.DoubleVal()
			}
			require.Equal(t, 3, len(bufferPoolSizeMetrics))
			require.Equal(t, map[string]float64{
				"mysql.buffer_pool_size :data":  16072704,
				"mysql.buffer_pool_size :dirty": 0,
				"mysql.buffer_pool_size :size":  134217728,
			}, bufferPoolSizeMetrics)
		case metadata.M.MysqlCommands.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 6, dps.Len())
			commandsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Command)
				label := fmt.Sprintf("%s :%s", m.Name(), value_label)
				commandsMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 6, len(commandsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.commands :close":          0,
				"mysql.commands :execute":        0,
				"mysql.commands :fetch":          0,
				"mysql.commands :prepare":        0,
				"mysql.commands :reset":          0,
				"mysql.commands :send_long_data": 0,
			}, commandsMetrics)
		case metadata.M.MysqlHandlers.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 18, dps.Len())
			handlersMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Handler)
				label := fmt.Sprintf("%s :%s", m.Name(), value_label)
				handlersMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 18, len(handlersMetrics))
			require.Equal(t, map[string]int64{
				"mysql.handlers :commit":             564,
				"mysql.handlers :delete":             0,
				"mysql.handlers :discover":           0,
				"mysql.handlers :lock":               7127,
				"mysql.handlers :mrr_init":           0,
				"mysql.handlers :prepare":            0,
				"mysql.handlers :read_first":         52,
				"mysql.handlers :read_key":           1680,
				"mysql.handlers :read_last":          0,
				"mysql.handlers :read_next":          3960,
				"mysql.handlers :read_prev":          0,
				"mysql.handlers :read_rnd":           0,
				"mysql.handlers :read_rnd_next":      505063,
				"mysql.handlers :rollback":           0,
				"mysql.handlers :savepoint":          0,
				"mysql.handlers :savepoint_rollback": 0,
				"mysql.handlers :update":             315,
				"mysql.handlers :write":              256734,
			}, handlersMetrics)
		case metadata.M.MysqlDoubleWrites.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			doubleWritesMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.DoubleWrites)
				label := fmt.Sprintf("%s :%s", m.Name(), value_label)
				doubleWritesMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 2, len(doubleWritesMetrics))
			require.Equal(t, map[string]int64{
				"mysql.double_writes :writes":  8,
				"mysql.double_writes :written": 27,
			}, doubleWritesMetrics)
		case metadata.M.MysqlLogOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			logOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.LogOperations)
				label := fmt.Sprintf("%s :%s", m.Name(), value_label)
				logOperationsMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 3, len(logOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.log_operations :requests": 646,
				"mysql.log_operations :waits":    0,
				"mysql.log_operations :writes":   14,
			}, logOperationsMetrics)
		case metadata.M.MysqlOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			operationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Operations)
				label := fmt.Sprintf("%s :%s", m.Name(), value_label)
				operationsMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 3, len(operationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.operations :fsyncs": 46,
				"mysql.operations :reads":  860,
				"mysql.operations :writes": 215,
			}, operationsMetrics)
		case metadata.M.MysqlPageOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			pageOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.PageOperations)
				label := fmt.Sprintf("%s :%s", m.Name(), value_label)
				pageOperationsMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 3, len(pageOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.page_operations :created": 144,
				"mysql.page_operations :read":    837,
				"mysql.page_operations :written": 168,
			}, pageOperationsMetrics)
		case metadata.M.MysqlRowLocks.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			rowLocksMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.RowLocks)
				label := fmt.Sprintf("%s :%s", m.Name(), value_label)
				rowLocksMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 2, len(rowLocksMetrics))
			require.Equal(t, map[string]int64{
				"mysql.row_locks :time":  0,
				"mysql.row_locks :waits": 0,
			}, rowLocksMetrics)
		case metadata.M.MysqlRowOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 4, dps.Len())
			rowOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.RowOperations)
				label := fmt.Sprintf("%s :%s", m.Name(), value_label)
				rowOperationsMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 4, len(rowOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.row_operations :deleted":  0,
				"mysql.row_operations :inserted": 0,
				"mysql.row_operations :read":     0,
				"mysql.row_operations :updated":  0,
			}, rowOperationsMetrics)
		case metadata.M.MysqlLocks.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			locksMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Locks)
				label := fmt.Sprintf("%s :%s", m.Name(), value_label)
				locksMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 2, len(locksMetrics))
			require.Equal(t, map[string]int64{
				"mysql.locks :immediate": 521,
				"mysql.locks :waited":    0,
			}, locksMetrics)
		case metadata.M.MysqlSorts.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 4, dps.Len())
			sortsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Sorts)
				label := fmt.Sprintf("%s :%s", m.Name(), value_label)
				sortsMetrics[label] = dp.IntVal()
			}
			require.Equal(t, 4, len(sortsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.sorts :merge_passes": 0,
				"mysql.sorts :range":        0,
				"mysql.sorts :rows":         0,
				"mysql.sorts :scan":         0,
			}, sortsMetrics)
		case metadata.M.MysqlThreads.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 4, dps.Len())
			threadsMetrics := map[string]float64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Threads)
				label := fmt.Sprintf("%s :%s", m.Name(), value_label)
				threadsMetrics[label] = dp.DoubleVal()
			}
			require.Equal(t, 4, len(threadsMetrics))
			require.Equal(t, map[string]float64{
				"mysql.threads :cached":    0,
				"mysql.threads :connected": 1,
				"mysql.threads :created":   1,
				"mysql.threads :running":   2,
			}, threadsMetrics)
		}
	}
}
