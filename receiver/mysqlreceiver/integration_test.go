package mysqlreceiver

import (
	"context"
	"fmt"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.uber.org/zap"

	"github.com/observiq/opentelemetry-components/receiver/mysqlreceiver/internal/metadata"
)

func mysqlContainer(t *testing.T) testcontainers.Container {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.mysql",
		},
		ExposedPorts: []string{"3306:3306"},
		WaitingFor:   wait.ForListeningPort("3306"),
	}
	require.NoError(t, req.Validate())

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	code, err := container.Exec(context.Background(), []string{"/setup.sh"})
	require.NoError(t, err)
	require.Equal(t, 0, code)

	return container
}

type MysqlIntegrationSuite struct {
	suite.Suite
}

func TestMysqlIntegration(t *testing.T) {
	suite.Run(t, new(MysqlIntegrationSuite))
}

func (suite *MysqlIntegrationSuite) TestHappyPathNoDatabase() {
	t := suite.T()
	container := mysqlContainer(t)
	defer func() {
		err := container.Terminate(context.Background())
		require.NoError(t, err)
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	sc := newMySQLScraper(zap.NewNop(), &Config{
		Username: "otel",
		Password: "otel",
		Endpoint: fmt.Sprintf("%s:3306", hostname),
	})
	err = sc.start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)
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
			bufferPoolPagesMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolPagesState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				bufferPoolPagesMetrics[label] = true
			}
			require.Equal(t, map[string]bool{
				"mysql.buffer_pool_pages state:data database:_global":    true,
				"mysql.buffer_pool_pages state:dirty database:_global":   true,
				"mysql.buffer_pool_pages state:flushed database:_global": true,
				"mysql.buffer_pool_pages state:free database:_global":    true,
				"mysql.buffer_pool_pages state:misc database:_global":    true,
				"mysql.buffer_pool_pages state:total database:_global":   true,
			}, bufferPoolPagesMetrics)
		case metadata.M.MysqlBufferPoolOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 7, dps.Len())
			bufferPoolOperationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolOperationsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				bufferPoolOperationsMetrics[label] = true
			}
			require.Equal(t, 7, len(bufferPoolOperationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.buffer_pool_operations state:read_ahead database:_global":         true,
				"mysql.buffer_pool_operations state:read_ahead_evicted database:_global": true,
				"mysql.buffer_pool_operations state:read_ahead_rnd database:_global":     true,
				"mysql.buffer_pool_operations state:read_requests database:_global":      true,
				"mysql.buffer_pool_operations state:reads database:_global":              true,
				"mysql.buffer_pool_operations state:wait_free database:_global":          true,
				"mysql.buffer_pool_operations state:write_requests database:_global":     true,
			}, bufferPoolOperationsMetrics)
		case metadata.M.MysqlBufferPoolSize.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 3, dps.Len())
			bufferPoolSizeMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolSizeState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				bufferPoolSizeMetrics[label] = true
			}
			require.Equal(t, 3, len(bufferPoolSizeMetrics))
			require.Equal(t, map[string]bool{
				"mysql.buffer_pool_size state:data database:_global":  true,
				"mysql.buffer_pool_size state:dirty database:_global": true,
				"mysql.buffer_pool_size state:size database:_global":  true,
			}, bufferPoolSizeMetrics)
		case metadata.M.MysqlCommands.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 6, dps.Len())
			commandsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.CommandState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				commandsMetrics[label] = true
			}
			require.Equal(t, 6, len(commandsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.commands state:close database:_global":          true,
				"mysql.commands state:execute database:_global":        true,
				"mysql.commands state:fetch database:_global":          true,
				"mysql.commands state:prepare database:_global":        true,
				"mysql.commands state:reset database:_global":          true,
				"mysql.commands state:send_long_data database:_global": true,
			}, commandsMetrics)
		case metadata.M.MysqlHandlers.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 18, dps.Len())
			handlersMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.HandlerState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				handlersMetrics[label] = true
			}
			require.Equal(t, 18, len(handlersMetrics))
			require.Equal(t, map[string]bool{
				"mysql.handlers state:commit database:_global":             true,
				"mysql.handlers state:delete database:_global":             true,
				"mysql.handlers state:discover database:_global":           true,
				"mysql.handlers state:lock database:_global":               true,
				"mysql.handlers state:mrr_init database:_global":           true,
				"mysql.handlers state:prepare database:_global":            true,
				"mysql.handlers state:read_first database:_global":         true,
				"mysql.handlers state:read_key database:_global":           true,
				"mysql.handlers state:read_last database:_global":          true,
				"mysql.handlers state:read_next database:_global":          true,
				"mysql.handlers state:read_prev database:_global":          true,
				"mysql.handlers state:read_rnd database:_global":           true,
				"mysql.handlers state:read_rnd_next database:_global":      true,
				"mysql.handlers state:rollback database:_global":           true,
				"mysql.handlers state:savepoint database:_global":          true,
				"mysql.handlers state:savepoint_rollback database:_global": true,
				"mysql.handlers state:update database:_global":             true,
				"mysql.handlers state:write database:_global":              true,
			}, handlersMetrics)
		case metadata.M.MysqlDoubleWrites.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			doubleWritesMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.DoubleWritesState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				doubleWritesMetrics[label] = true
			}
			require.Equal(t, 2, len(doubleWritesMetrics))
			require.Equal(t, map[string]bool{
				"mysql.double_writes state:writes database:_global":  true,
				"mysql.double_writes state:written database:_global": true,
			}, doubleWritesMetrics)
		case metadata.M.MysqlLogOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			logOperationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.LogOperationsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				logOperationsMetrics[label] = true
			}
			require.Equal(t, 3, len(logOperationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.log_operations state:requests database:_global": true,
				"mysql.log_operations state:waits database:_global":    true,
				"mysql.log_operations state:writes database:_global":   true,
			}, logOperationsMetrics)
		case metadata.M.MysqlOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			operationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.OperationsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				operationsMetrics[label] = true
			}
			require.Equal(t, 3, len(operationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.operations state:fsyncs database:_global": true,
				"mysql.operations state:reads database:_global":  true,
				"mysql.operations state:writes database:_global": true,
			}, operationsMetrics)
		case metadata.M.MysqlPageOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			pageOperationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.PageOperationsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				pageOperationsMetrics[label] = true
			}
			require.Equal(t, 3, len(pageOperationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.page_operations state:created database:_global": true,
				"mysql.page_operations state:read database:_global":    true,
				"mysql.page_operations state:written database:_global": true,
			}, pageOperationsMetrics)
		case metadata.M.MysqlRowLocks.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			rowLocksMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.RowLocksState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				rowLocksMetrics[label] = true
			}
			require.Equal(t, 2, len(rowLocksMetrics))
			require.Equal(t, map[string]bool{
				"mysql.row_locks state:time database:_global":  true,
				"mysql.row_locks state:waits database:_global": true,
			}, rowLocksMetrics)
		case metadata.M.MysqlRowOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 4, dps.Len())
			rowOperationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.RowOperationsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				rowOperationsMetrics[label] = true
			}
			require.Equal(t, 4, len(rowOperationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.row_operations state:deleted database:_global":  true,
				"mysql.row_operations state:inserted database:_global": true,
				"mysql.row_operations state:read database:_global":     true,
				"mysql.row_operations state:updated database:_global":  true,
			}, rowOperationsMetrics)
		case metadata.M.MysqlLocks.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			locksMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.LocksState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				locksMetrics[label] = true
			}
			require.Equal(t, 2, len(locksMetrics))
			require.Equal(t, map[string]bool{
				"mysql.locks state:immediate database:_global": true,
				"mysql.locks state:waited database:_global":    true,
			}, locksMetrics)
		case metadata.M.MysqlSorts.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 4, dps.Len())
			sortsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.SortsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				sortsMetrics[label] = true
			}
			require.Equal(t, 4, len(sortsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.sorts state:merge_passes database:_global": true,
				"mysql.sorts state:range database:_global":        true,
				"mysql.sorts state:rows database:_global":         true,
				"mysql.sorts state:scan database:_global":         true,
			}, sortsMetrics)
		case metadata.M.MysqlThreads.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 4, dps.Len())
			threadsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.ThreadsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				threadsMetrics[label] = true
			}
			require.Equal(t, 4, len(threadsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.threads state:cached database:_global":    true,
				"mysql.threads state:connected database:_global": true,
				"mysql.threads state:created database:_global":   true,
				"mysql.threads state:running database:_global":   true,
			}, threadsMetrics)
		}
	}
}

func (suite *MysqlIntegrationSuite) TestHappyPathWithDatabase() {
	t := suite.T()
	container := mysqlContainer(t)
	defer func() {
		err := container.Terminate(context.Background())
		require.NoError(t, err)
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	sc := newMySQLScraper(zap.NewNop(), &Config{
		Username: "otel",
		Password: "otel",
		Database: "otel",
		Endpoint: fmt.Sprintf("%s:3306", hostname),
	})
	err = sc.start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)
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
			bufferPoolPagesMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolPagesState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				bufferPoolPagesMetrics[label] = true
			}
			require.Equal(t, map[string]bool{
				"mysql.buffer_pool_pages state:data database:otel":    true,
				"mysql.buffer_pool_pages state:dirty database:otel":   true,
				"mysql.buffer_pool_pages state:flushed database:otel": true,
				"mysql.buffer_pool_pages state:free database:otel":    true,
				"mysql.buffer_pool_pages state:misc database:otel":    true,
				"mysql.buffer_pool_pages state:total database:otel":   true,
			}, bufferPoolPagesMetrics)
		case metadata.M.MysqlBufferPoolOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 7, dps.Len())
			bufferPoolOperationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolOperationsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				bufferPoolOperationsMetrics[label] = true
			}
			require.Equal(t, 7, len(bufferPoolOperationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.buffer_pool_operations state:read_ahead database:otel":         true,
				"mysql.buffer_pool_operations state:read_ahead_evicted database:otel": true,
				"mysql.buffer_pool_operations state:read_ahead_rnd database:otel":     true,
				"mysql.buffer_pool_operations state:read_requests database:otel":      true,
				"mysql.buffer_pool_operations state:reads database:otel":              true,
				"mysql.buffer_pool_operations state:wait_free database:otel":          true,
				"mysql.buffer_pool_operations state:write_requests database:otel":     true,
			}, bufferPoolOperationsMetrics)
		case metadata.M.MysqlBufferPoolSize.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 3, dps.Len())
			bufferPoolSizeMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolSizeState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				bufferPoolSizeMetrics[label] = true
			}
			require.Equal(t, 3, len(bufferPoolSizeMetrics))
			require.Equal(t, map[string]bool{
				"mysql.buffer_pool_size state:data database:otel":  true,
				"mysql.buffer_pool_size state:dirty database:otel": true,
				"mysql.buffer_pool_size state:size database:otel":  true,
			}, bufferPoolSizeMetrics)
		case metadata.M.MysqlCommands.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 6, dps.Len())
			commandsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.CommandState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				commandsMetrics[label] = true
			}
			require.Equal(t, 6, len(commandsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.commands state:close database:otel":          true,
				"mysql.commands state:execute database:otel":        true,
				"mysql.commands state:fetch database:otel":          true,
				"mysql.commands state:prepare database:otel":        true,
				"mysql.commands state:reset database:otel":          true,
				"mysql.commands state:send_long_data database:otel": true,
			}, commandsMetrics)
		case metadata.M.MysqlHandlers.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 18, dps.Len())
			handlersMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.HandlerState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				handlersMetrics[label] = true
			}
			require.Equal(t, 18, len(handlersMetrics))
			require.Equal(t, map[string]bool{
				"mysql.handlers state:commit database:otel":             true,
				"mysql.handlers state:delete database:otel":             true,
				"mysql.handlers state:discover database:otel":           true,
				"mysql.handlers state:lock database:otel":               true,
				"mysql.handlers state:mrr_init database:otel":           true,
				"mysql.handlers state:prepare database:otel":            true,
				"mysql.handlers state:read_first database:otel":         true,
				"mysql.handlers state:read_key database:otel":           true,
				"mysql.handlers state:read_last database:otel":          true,
				"mysql.handlers state:read_next database:otel":          true,
				"mysql.handlers state:read_prev database:otel":          true,
				"mysql.handlers state:read_rnd database:otel":           true,
				"mysql.handlers state:read_rnd_next database:otel":      true,
				"mysql.handlers state:rollback database:otel":           true,
				"mysql.handlers state:savepoint database:otel":          true,
				"mysql.handlers state:savepoint_rollback database:otel": true,
				"mysql.handlers state:update database:otel":             true,
				"mysql.handlers state:write database:otel":              true,
			}, handlersMetrics)
		case metadata.M.MysqlDoubleWrites.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			doubleWritesMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.DoubleWritesState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				doubleWritesMetrics[label] = true
			}
			require.Equal(t, 2, len(doubleWritesMetrics))
			require.Equal(t, map[string]bool{
				"mysql.double_writes state:writes database:otel":  true,
				"mysql.double_writes state:written database:otel": true,
			}, doubleWritesMetrics)
		case metadata.M.MysqlLogOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			logOperationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.LogOperationsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				logOperationsMetrics[label] = true
			}
			require.Equal(t, 3, len(logOperationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.log_operations state:requests database:otel": true,
				"mysql.log_operations state:waits database:otel":    true,
				"mysql.log_operations state:writes database:otel":   true,
			}, logOperationsMetrics)
		case metadata.M.MysqlOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			operationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.OperationsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				operationsMetrics[label] = true
			}
			require.Equal(t, 3, len(operationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.operations state:fsyncs database:otel": true,
				"mysql.operations state:reads database:otel":  true,
				"mysql.operations state:writes database:otel": true,
			}, operationsMetrics)
		case metadata.M.MysqlPageOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			pageOperationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.PageOperationsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				pageOperationsMetrics[label] = true
			}
			require.Equal(t, 3, len(pageOperationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.page_operations state:created database:otel": true,
				"mysql.page_operations state:read database:otel":    true,
				"mysql.page_operations state:written database:otel": true,
			}, pageOperationsMetrics)
		case metadata.M.MysqlRowLocks.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			rowLocksMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.RowLocksState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				rowLocksMetrics[label] = true
			}
			require.Equal(t, 2, len(rowLocksMetrics))
			require.Equal(t, map[string]bool{
				"mysql.row_locks state:time database:otel":  true,
				"mysql.row_locks state:waits database:otel": true,
			}, rowLocksMetrics)
		case metadata.M.MysqlRowOperations.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 4, dps.Len())
			rowOperationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.RowOperationsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				rowOperationsMetrics[label] = true
			}
			require.Equal(t, 4, len(rowOperationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.row_operations state:deleted database:otel":  true,
				"mysql.row_operations state:inserted database:otel": true,
				"mysql.row_operations state:read database:otel":     true,
				"mysql.row_operations state:updated database:otel":  true,
			}, rowOperationsMetrics)
		case metadata.M.MysqlLocks.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			locksMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.LocksState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				locksMetrics[label] = true
			}
			require.Equal(t, 2, len(locksMetrics))
			require.Equal(t, map[string]bool{
				"mysql.locks state:immediate database:otel": true,
				"mysql.locks state:waited database:otel":    true,
			}, locksMetrics)
		case metadata.M.MysqlSorts.Name():
			dps := m.IntSum().DataPoints()
			require.True(t, m.IntSum().IsMonotonic())
			require.Equal(t, 4, dps.Len())
			sortsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.SortsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				sortsMetrics[label] = true
			}
			require.Equal(t, 4, len(sortsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.sorts state:merge_passes database:otel": true,
				"mysql.sorts state:range database:otel":        true,
				"mysql.sorts state:rows database:otel":         true,
				"mysql.sorts state:scan database:otel":         true,
			}, sortsMetrics)
		case metadata.M.MysqlThreads.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 4, dps.Len())
			threadsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.ThreadsState)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s state:%s database:%s", m.Name(), value_label, db_label)
				threadsMetrics[label] = true
			}
			require.Equal(t, 4, len(threadsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.threads state:cached database:otel":    true,
				"mysql.threads state:connected database:otel": true,
				"mysql.threads state:created database:otel":   true,
				"mysql.threads state:running database:otel":   true,
			}, threadsMetrics)
		}
	}
}

func (suite *MysqlIntegrationSuite) TestStartStop() {
	t := suite.T()
	container := mysqlContainer(t)
	defer func() {
		err := container.Terminate(context.Background())
		require.NoError(t, err)
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	sc := newMySQLScraper(zap.NewNop(), &Config{
		Username: "otel",
		Password: "otel",
		Database: "otel",
		Endpoint: fmt.Sprintf("%s:3306", hostname),
	})

	// scraper is connected
	err = sc.start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	// scraper is closed
	err = sc.shutdown(context.Background())
	require.NoError(t, err)

	// scraper scapes and produces an error because db is closed
	_, err = sc.scrape(context.Background())
	require.Error(t, err)
	require.EqualError(t, err, "sql: database is closed")
}
