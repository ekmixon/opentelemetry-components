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
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolPagesState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				bufferPoolPagesMetrics[label] = dp.Value()
			}
			require.Equal(t, 6, len(bufferPoolPagesMetrics))
			require.Equal(t, map[string]float64{
				"mysql.buffer_pool_pages state:data database:otel":    981,
				"mysql.buffer_pool_pages state:dirty database:otel":   0,
				"mysql.buffer_pool_pages state:flushed database:otel": 168,
				"mysql.buffer_pool_pages state:free database:otel":    7207,
				"mysql.buffer_pool_pages state:misc database:otel":    4,
				"mysql.buffer_pool_pages state:total database:otel":   8192,
			}, bufferPoolPagesMetrics)
		case metadata.M.MysqlBufferPoolOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 7, dps.Len())
			bufferPoolOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolOperationsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				bufferPoolOperationsMetrics[label] = dp.Value()
			}
			require.Equal(t, 7, len(bufferPoolOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.buffer_pool_operations state:read_ahead database:otel":         0,
				"mysql.buffer_pool_operations state:read_ahead_evicted database:otel": 0,
				"mysql.buffer_pool_operations state:read_ahead_rnd database:otel":     0,
				"mysql.buffer_pool_operations state:read_requests database:otel":      14837,
				"mysql.buffer_pool_operations state:reads database:otel":              838,
				"mysql.buffer_pool_operations state:wait_free database:otel":          0,
				"mysql.buffer_pool_operations state:write_requests database:otel":     1669,
			}, bufferPoolOperationsMetrics)
		case metadata.M.MysqlBufferPoolSize.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 3, dps.Len())
			bufferPoolSizeMetrics := map[string]float64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolSizeState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				bufferPoolSizeMetrics[label] = dp.Value()
			}
			require.Equal(t, 3, len(bufferPoolSizeMetrics))
			require.Equal(t, map[string]float64{
				"mysql.buffer_pool_size state:data database:otel":  16072704,
				"mysql.buffer_pool_size state:dirty database:otel": 0,
				"mysql.buffer_pool_size state:size database:otel":  134217728,
			}, bufferPoolSizeMetrics)
		case metadata.M.MysqlCommands.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 6, dps.Len())
			commandsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.CommandState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				commandsMetrics[label] = dp.Value()
			}
			require.Equal(t, 6, len(commandsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.commands state:close database:otel":          0,
				"mysql.commands state:execute database:otel":        0,
				"mysql.commands state:fetch database:otel":          0,
				"mysql.commands state:prepare database:otel":        0,
				"mysql.commands state:reset database:otel":          0,
				"mysql.commands state:send_long_data database:otel": 0,
			}, commandsMetrics)
		case metadata.M.MysqlHandlers.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 18, dps.Len())
			handlersMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.HandlerState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				handlersMetrics[label] = dp.Value()
			}
			require.Equal(t, 18, len(handlersMetrics))
			require.Equal(t, map[string]int64{
				"mysql.handlers state:commit database:otel":             564,
				"mysql.handlers state:delete database:otel":             0,
				"mysql.handlers state:discover database:otel":           0,
				"mysql.handlers state:lock database:otel":               7127,
				"mysql.handlers state:mrr_init database:otel":           0,
				"mysql.handlers state:prepare database:otel":            0,
				"mysql.handlers state:read_first database:otel":         52,
				"mysql.handlers state:read_key database:otel":           1680,
				"mysql.handlers state:read_last database:otel":          0,
				"mysql.handlers state:read_next database:otel":          3960,
				"mysql.handlers state:read_prev database:otel":          0,
				"mysql.handlers state:read_rnd database:otel":           0,
				"mysql.handlers state:read_rnd_next database:otel":      505063,
				"mysql.handlers state:rollback database:otel":           0,
				"mysql.handlers state:savepoint database:otel":          0,
				"mysql.handlers state:savepoint_rollback database:otel": 0,
				"mysql.handlers state:update database:otel":             315,
				"mysql.handlers state:write database:otel":              256734,
			}, handlersMetrics)
		case metadata.M.MysqlDoubleWrites.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			doubleWritesMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.DoubleWritesState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				doubleWritesMetrics[label] = dp.Value()
			}
			require.Equal(t, 2, len(doubleWritesMetrics))
			require.Equal(t, map[string]int64{
				"mysql.double_writes state:writes database:otel":  8,
				"mysql.double_writes state:written database:otel": 27,
			}, doubleWritesMetrics)
		case metadata.M.MysqlLogOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			logOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.LogOperationsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				logOperationsMetrics[label] = dp.Value()
			}
			require.Equal(t, 3, len(logOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.log_operations state:requests database:otel": 646,
				"mysql.log_operations state:waits database:otel":    0,
				"mysql.log_operations state:writes database:otel":   14,
			}, logOperationsMetrics)
		case metadata.M.MysqlOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			operationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.OperationsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				operationsMetrics[label] = dp.Value()
			}
			require.Equal(t, 3, len(operationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.operations state:fsyncs database:otel": 46,
				"mysql.operations state:reads database:otel":  860,
				"mysql.operations state:writes database:otel": 215,
			}, operationsMetrics)
		case metadata.M.MysqlPageOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			pageOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.PageOperationsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				pageOperationsMetrics[label] = dp.Value()
			}
			require.Equal(t, 3, len(pageOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.page_operations state:created database:otel": 144,
				"mysql.page_operations state:read database:otel":    837,
				"mysql.page_operations state:written database:otel": 168,
			}, pageOperationsMetrics)
		case metadata.M.MysqlRowLocks.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			rowLocksMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.RowLocksState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				rowLocksMetrics[label] = dp.Value()
			}
			require.Equal(t, 2, len(rowLocksMetrics))
			require.Equal(t, map[string]int64{
				"mysql.row_locks state:time database:otel":  0,
				"mysql.row_locks state:waits database:otel": 0,
			}, rowLocksMetrics)
		case metadata.M.MysqlRowOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 4, dps.Len())
			rowOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.RowOperationsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				rowOperationsMetrics[label] = dp.Value()
			}
			require.Equal(t, 4, len(rowOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.row_operations state:deleted database:otel":  0,
				"mysql.row_operations state:inserted database:otel": 0,
				"mysql.row_operations state:read database:otel":     0,
				"mysql.row_operations state:updated database:otel":  0,
			}, rowOperationsMetrics)
		case metadata.M.MysqlLocks.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			locksMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.LocksState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				locksMetrics[label] = dp.Value()
			}
			require.Equal(t, 2, len(locksMetrics))
			require.Equal(t, map[string]int64{
				"mysql.locks state:immediate database:otel": 521,
				"mysql.locks state:waited database:otel":    0,
			}, locksMetrics)
		case metadata.M.MysqlSorts.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 4, dps.Len())
			sortsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.SortsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				sortsMetrics[label] = dp.Value()
			}
			require.Equal(t, 4, len(sortsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.sorts state:merge_passes database:otel": 0,
				"mysql.sorts state:range database:otel":        0,
				"mysql.sorts state:rows database:otel":         0,
				"mysql.sorts state:scan database:otel":         0,
			}, sortsMetrics)
		case metadata.M.MysqlThreads.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 4, dps.Len())
			threadsMetrics := map[string]float64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.ThreadsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				threadsMetrics[label] = dp.Value()
			}
			require.Equal(t, 4, len(threadsMetrics))
			require.Equal(t, map[string]float64{
				"mysql.threads state:cached database:otel":    0,
				"mysql.threads state:connected database:otel": 1,
				"mysql.threads state:created database:otel":   1,
				"mysql.threads state:running database:otel":   2,
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
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolPagesState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				bufferPoolPagesMetrics[label] = dp.Value()
			}
			require.Equal(t, 6, len(bufferPoolPagesMetrics))
			require.Equal(t, map[string]float64{
				"mysql.buffer_pool_pages state:data database:_global":    981,
				"mysql.buffer_pool_pages state:dirty database:_global":   0,
				"mysql.buffer_pool_pages state:flushed database:_global": 168,
				"mysql.buffer_pool_pages state:free database:_global":    7207,
				"mysql.buffer_pool_pages state:misc database:_global":    4,
				"mysql.buffer_pool_pages state:total database:_global":   8192,
			}, bufferPoolPagesMetrics)
		case metadata.M.MysqlBufferPoolOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 7, dps.Len())
			bufferPoolOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolOperationsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				bufferPoolOperationsMetrics[label] = dp.Value()
			}
			require.Equal(t, 7, len(bufferPoolOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.buffer_pool_operations state:read_ahead database:_global":         0,
				"mysql.buffer_pool_operations state:read_ahead_evicted database:_global": 0,
				"mysql.buffer_pool_operations state:read_ahead_rnd database:_global":     0,
				"mysql.buffer_pool_operations state:read_requests database:_global":      14837,
				"mysql.buffer_pool_operations state:reads database:_global":              838,
				"mysql.buffer_pool_operations state:wait_free database:_global":          0,
				"mysql.buffer_pool_operations state:write_requests database:_global":     1669,
			}, bufferPoolOperationsMetrics)
		case metadata.M.MysqlBufferPoolSize.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 3, dps.Len())
			bufferPoolSizeMetrics := map[string]float64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolSizeState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				bufferPoolSizeMetrics[label] = dp.Value()
			}
			require.Equal(t, 3, len(bufferPoolSizeMetrics))
			require.Equal(t, map[string]float64{
				"mysql.buffer_pool_size state:data database:_global":  16072704,
				"mysql.buffer_pool_size state:dirty database:_global": 0,
				"mysql.buffer_pool_size state:size database:_global":  134217728,
			}, bufferPoolSizeMetrics)
		case metadata.M.MysqlCommands.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 6, dps.Len())
			commandsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.CommandState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				commandsMetrics[label] = dp.Value()
			}
			require.Equal(t, 6, len(commandsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.commands state:close database:_global":          0,
				"mysql.commands state:execute database:_global":        0,
				"mysql.commands state:fetch database:_global":          0,
				"mysql.commands state:prepare database:_global":        0,
				"mysql.commands state:reset database:_global":          0,
				"mysql.commands state:send_long_data database:_global": 0,
			}, commandsMetrics)
		case metadata.M.MysqlHandlers.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 18, dps.Len())
			handlersMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.HandlerState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				handlersMetrics[label] = dp.Value()
			}
			require.Equal(t, 18, len(handlersMetrics))
			require.Equal(t, map[string]int64{
				"mysql.handlers state:commit database:_global":             564,
				"mysql.handlers state:delete database:_global":             0,
				"mysql.handlers state:discover database:_global":           0,
				"mysql.handlers state:lock database:_global":               7127,
				"mysql.handlers state:mrr_init database:_global":           0,
				"mysql.handlers state:prepare database:_global":            0,
				"mysql.handlers state:read_first database:_global":         52,
				"mysql.handlers state:read_key database:_global":           1680,
				"mysql.handlers state:read_last database:_global":          0,
				"mysql.handlers state:read_next database:_global":          3960,
				"mysql.handlers state:read_prev database:_global":          0,
				"mysql.handlers state:read_rnd database:_global":           0,
				"mysql.handlers state:read_rnd_next database:_global":      505063,
				"mysql.handlers state:rollback database:_global":           0,
				"mysql.handlers state:savepoint database:_global":          0,
				"mysql.handlers state:savepoint_rollback database:_global": 0,
				"mysql.handlers state:update database:_global":             315,
				"mysql.handlers state:write database:_global":              256734,
			}, handlersMetrics)
		case metadata.M.MysqlDoubleWrites.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			doubleWritesMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.DoubleWritesState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				doubleWritesMetrics[label] = dp.Value()
			}
			require.Equal(t, 2, len(doubleWritesMetrics))
			require.Equal(t, map[string]int64{
				"mysql.double_writes state:writes database:_global":  8,
				"mysql.double_writes state:written database:_global": 27,
			}, doubleWritesMetrics)
		case metadata.M.MysqlLogOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			logOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.LogOperationsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				logOperationsMetrics[label] = dp.Value()
			}
			require.Equal(t, 3, len(logOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.log_operations state:requests database:_global": 646,
				"mysql.log_operations state:waits database:_global":    0,
				"mysql.log_operations state:writes database:_global":   14,
			}, logOperationsMetrics)
		case metadata.M.MysqlOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			operationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.OperationsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				operationsMetrics[label] = dp.Value()
			}
			require.Equal(t, 3, len(operationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.operations state:fsyncs database:_global": 46,
				"mysql.operations state:reads database:_global":  860,
				"mysql.operations state:writes database:_global": 215,
			}, operationsMetrics)
		case metadata.M.MysqlPageOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			pageOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.PageOperationsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				pageOperationsMetrics[label] = dp.Value()
			}
			require.Equal(t, 3, len(pageOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.page_operations state:created database:_global": 144,
				"mysql.page_operations state:read database:_global":    837,
				"mysql.page_operations state:written database:_global": 168,
			}, pageOperationsMetrics)
		case metadata.M.MysqlRowLocks.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			rowLocksMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.RowLocksState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				rowLocksMetrics[label] = dp.Value()
			}
			require.Equal(t, 2, len(rowLocksMetrics))
			require.Equal(t, map[string]int64{
				"mysql.row_locks state:time database:_global":  0,
				"mysql.row_locks state:waits database:_global": 0,
			}, rowLocksMetrics)
		case metadata.M.MysqlRowOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 4, dps.Len())
			rowOperationsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.RowOperationsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				rowOperationsMetrics[label] = dp.Value()
			}
			require.Equal(t, 4, len(rowOperationsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.row_operations state:deleted database:_global":  0,
				"mysql.row_operations state:inserted database:_global": 0,
				"mysql.row_operations state:read database:_global":     0,
				"mysql.row_operations state:updated database:_global":  0,
			}, rowOperationsMetrics)
		case metadata.M.MysqlLocks.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			locksMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.LocksState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				locksMetrics[label] = dp.Value()
			}
			require.Equal(t, 2, len(locksMetrics))
			require.Equal(t, map[string]int64{
				"mysql.locks state:immediate database:_global": 521,
				"mysql.locks state:waited database:_global":    0,
			}, locksMetrics)
		case metadata.M.MysqlSorts.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 4, dps.Len())
			sortsMetrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.SortsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				sortsMetrics[label] = dp.Value()
			}
			require.Equal(t, 4, len(sortsMetrics))
			require.Equal(t, map[string]int64{
				"mysql.sorts state:merge_passes database:_global": 0,
				"mysql.sorts state:range database:_global":        0,
				"mysql.sorts state:rows database:_global":         0,
				"mysql.sorts state:scan database:_global":         0,
			}, sortsMetrics)
		case metadata.M.MysqlThreads.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 4, dps.Len())
			threadsMetrics := map[string]float64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.ThreadsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				threadsMetrics[label] = dp.Value()
			}
			require.Equal(t, 4, len(threadsMetrics))
			require.Equal(t, map[string]float64{
				"mysql.threads state:cached database:_global":    0,
				"mysql.threads state:connected database:_global": 1,
				"mysql.threads state:created database:_global":   1,
				"mysql.threads state:running database:_global":   2,
			}, threadsMetrics)
		}
	}
}
