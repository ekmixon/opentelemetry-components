// +build integration

package mysqlreceiver

import (
	"context"
	"fmt"
	"net"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"

	"github.com/observiq/opentelemetry-components/receiver/mysqlreceiver/internal/metadata"
)

func TestMysqlIntegrationNoDatabase(t *testing.T) {
	container := getContainer(t, containerRequest8_0)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = net.JoinHostPort(hostname, "3306")
	cfg.Username = "otel"
	cfg.Password = "otel"

	consumer := new(consumertest.MetricsSink)
	settings := componenttest.NewNopReceiverCreateSettings()
	rcvr, err := f.CreateMetricsReceiver(context.Background(), settings, cfg, consumer)
	require.NoError(t, err, "failed creating metrics receiver")
	require.NoError(t, rcvr.Start(context.Background(), componenttest.NewNopHost()))
	require.Eventuallyf(t, func() bool {
		return len(consumer.AllMetrics()) > 0
	}, 2*time.Minute, 1*time.Second, "failed to receive more than 0 metrics")

	md := consumer.AllMetrics()[0]
	require.Equal(t, 1, md.ResourceMetrics().Len())
	ilms := md.ResourceMetrics().At(0).InstrumentationLibraryMetrics()
	require.Equal(t, 1, ilms.Len())
	metrics := ilms.At(0).Metrics()
	require.NoError(t, rcvr.Shutdown(context.Background()))

	validateNoDatabaseResult(t, metrics)
}

func TestMysqlIntegrationWithDatabase(t *testing.T) {
	container := getContainer(t, containerRequest8_0)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = fmt.Sprintf("%s:3306", hostname)
	cfg.Username = "otel"
	cfg.Password = "otel"
	cfg.Database = "otel"

	consumer := new(consumertest.MetricsSink)
	settings := componenttest.NewNopReceiverCreateSettings()
	rcvr, err := f.CreateMetricsReceiver(context.Background(), settings, cfg, consumer)
	require.NoError(t, err, "failed creating metrics receiver")
	require.NoError(t, rcvr.Start(context.Background(), componenttest.NewNopHost()))
	require.Eventuallyf(t, func() bool {
		return len(consumer.AllMetrics()) > 0
	}, 2*time.Minute, 1*time.Second, "failed to receive more than 0 metrics")

	md := consumer.AllMetrics()[0]
	require.Equal(t, 1, md.ResourceMetrics().Len())
	ilms := md.ResourceMetrics().At(0).InstrumentationLibraryMetrics()
	require.Equal(t, 1, ilms.Len())
	metrics := ilms.At(0).Metrics()
	require.NoError(t, rcvr.Shutdown(context.Background()))

	validateWithDatabaseResult(t, metrics)
}

var (
	containerRequest8_0 = testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.mysql",
		},
		ExposedPorts: []string{"3306:3306"},
		WaitingFor: wait.ForListeningPort("3306").
			WithStartupTimeout(2 * time.Minute),
	}
)

func getContainer(t *testing.T, req testcontainers.ContainerRequest) testcontainers.Container {
	require.NoError(t, req.Validate())
	container, err := testcontainers.GenericContainer(
		context.Background(),
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
	require.NoError(t, err)

	code, err := container.Exec(context.Background(), []string{"/setup.sh"})
	require.NoError(t, err)
	require.Equal(t, 0, code)
	return container
}

func validateNoDatabaseResult(t *testing.T, metrics pdata.MetricSlice) {
	require.Equal(t, len(metadata.M.Names()), metrics.Len())

	for i := 0; i < metrics.Len(); i++ {
		m := metrics.At(i)
		switch m.Name() {
		case metadata.M.MysqlBufferPoolPages.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 6, dps.Len())
			bufferPoolPagesMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolPages)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				bufferPoolPagesMetrics[label] = true
			}
			require.Equal(t, map[string]bool{
				"mysql.buffer_pool_pages :data database:_global":    true,
				"mysql.buffer_pool_pages :dirty database:_global":   true,
				"mysql.buffer_pool_pages :flushed database:_global": true,
				"mysql.buffer_pool_pages :free database:_global":    true,
				"mysql.buffer_pool_pages :misc database:_global":    true,
				"mysql.buffer_pool_pages :total database:_global":   true,
			}, bufferPoolPagesMetrics)
		case metadata.M.MysqlBufferPoolOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 7, dps.Len())
			bufferPoolOperationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolOperations)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				bufferPoolOperationsMetrics[label] = true
			}
			require.Equal(t, 7, len(bufferPoolOperationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.buffer_pool_operations :read_ahead database:_global":         true,
				"mysql.buffer_pool_operations :read_ahead_evicted database:_global": true,
				"mysql.buffer_pool_operations :read_ahead_rnd database:_global":     true,
				"mysql.buffer_pool_operations :read_requests database:_global":      true,
				"mysql.buffer_pool_operations :reads database:_global":              true,
				"mysql.buffer_pool_operations :wait_free database:_global":          true,
				"mysql.buffer_pool_operations :write_requests database:_global":     true,
			}, bufferPoolOperationsMetrics)
		case metadata.M.MysqlBufferPoolSize.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 3, dps.Len())
			bufferPoolSizeMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolSize)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				bufferPoolSizeMetrics[label] = true
			}
			require.Equal(t, 3, len(bufferPoolSizeMetrics))
			require.Equal(t, map[string]bool{
				"mysql.buffer_pool_size :data database:_global":  true,
				"mysql.buffer_pool_size :dirty database:_global": true,
				"mysql.buffer_pool_size :size database:_global":  true,
			}, bufferPoolSizeMetrics)
		case metadata.M.MysqlCommands.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 6, dps.Len())
			commandsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Command)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				commandsMetrics[label] = true
			}
			require.Equal(t, 6, len(commandsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.commands :close database:_global":          true,
				"mysql.commands :execute database:_global":        true,
				"mysql.commands :fetch database:_global":          true,
				"mysql.commands :prepare database:_global":        true,
				"mysql.commands :reset database:_global":          true,
				"mysql.commands :send_long_data database:_global": true,
			}, commandsMetrics)
		case metadata.M.MysqlHandlers.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 18, dps.Len())
			handlersMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Handler)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				handlersMetrics[label] = true
			}
			require.Equal(t, 18, len(handlersMetrics))
			require.Equal(t, map[string]bool{
				"mysql.handlers :commit database:_global":             true,
				"mysql.handlers :delete database:_global":             true,
				"mysql.handlers :discover database:_global":           true,
				"mysql.handlers :lock database:_global":               true,
				"mysql.handlers :mrr_init database:_global":           true,
				"mysql.handlers :prepare database:_global":            true,
				"mysql.handlers :read_first database:_global":         true,
				"mysql.handlers :read_key database:_global":           true,
				"mysql.handlers :read_last database:_global":          true,
				"mysql.handlers :read_next database:_global":          true,
				"mysql.handlers :read_prev database:_global":          true,
				"mysql.handlers :read_rnd database:_global":           true,
				"mysql.handlers :read_rnd_next database:_global":      true,
				"mysql.handlers :rollback database:_global":           true,
				"mysql.handlers :savepoint database:_global":          true,
				"mysql.handlers :savepoint_rollback database:_global": true,
				"mysql.handlers :update database:_global":             true,
				"mysql.handlers :write database:_global":              true,
			}, handlersMetrics)
		case metadata.M.MysqlDoubleWrites.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			doubleWritesMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.DoubleWrites)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				doubleWritesMetrics[label] = true
			}
			require.Equal(t, 2, len(doubleWritesMetrics))
			require.Equal(t, map[string]bool{
				"mysql.double_writes :writes database:_global":  true,
				"mysql.double_writes :written database:_global": true,
			}, doubleWritesMetrics)
		case metadata.M.MysqlLogOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			logOperationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.LogOperations)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				logOperationsMetrics[label] = true
			}
			require.Equal(t, 3, len(logOperationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.log_operations :requests database:_global": true,
				"mysql.log_operations :waits database:_global":    true,
				"mysql.log_operations :writes database:_global":   true,
			}, logOperationsMetrics)
		case metadata.M.MysqlOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			operationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Operations)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				operationsMetrics[label] = true
			}
			require.Equal(t, 3, len(operationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.operations :fsyncs database:_global": true,
				"mysql.operations :reads database:_global":  true,
				"mysql.operations :writes database:_global": true,
			}, operationsMetrics)
		case metadata.M.MysqlPageOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			pageOperationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.PageOperations)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				pageOperationsMetrics[label] = true
			}
			require.Equal(t, 3, len(pageOperationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.page_operations :created database:_global": true,
				"mysql.page_operations :read database:_global":    true,
				"mysql.page_operations :written database:_global": true,
			}, pageOperationsMetrics)
		case metadata.M.MysqlRowLocks.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			rowLocksMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.RowLocks)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				rowLocksMetrics[label] = true
			}
			require.Equal(t, 2, len(rowLocksMetrics))
			require.Equal(t, map[string]bool{
				"mysql.row_locks :time database:_global":  true,
				"mysql.row_locks :waits database:_global": true,
			}, rowLocksMetrics)
		case metadata.M.MysqlRowOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 4, dps.Len())
			rowOperationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.RowOperations)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				rowOperationsMetrics[label] = true
			}
			require.Equal(t, 4, len(rowOperationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.row_operations :deleted database:_global":  true,
				"mysql.row_operations :inserted database:_global": true,
				"mysql.row_operations :read database:_global":     true,
				"mysql.row_operations :updated database:_global":  true,
			}, rowOperationsMetrics)
		case metadata.M.MysqlLocks.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			locksMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Locks)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				locksMetrics[label] = true
			}
			require.Equal(t, 2, len(locksMetrics))
			require.Equal(t, map[string]bool{
				"mysql.locks :immediate database:_global": true,
				"mysql.locks :waited database:_global":    true,
			}, locksMetrics)
		case metadata.M.MysqlSorts.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 4, dps.Len())
			sortsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Sorts)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				sortsMetrics[label] = true
			}
			require.Equal(t, 4, len(sortsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.sorts :merge_passes database:_global": true,
				"mysql.sorts :range database:_global":        true,
				"mysql.sorts :rows database:_global":         true,
				"mysql.sorts :scan database:_global":         true,
			}, sortsMetrics)
		case metadata.M.MysqlThreads.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 4, dps.Len())
			threadsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Threads)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				threadsMetrics[label] = true
			}
			require.Equal(t, 4, len(threadsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.threads :cached database:_global":    true,
				"mysql.threads :connected database:_global": true,
				"mysql.threads :created database:_global":   true,
				"mysql.threads :running database:_global":   true,
			}, threadsMetrics)
		}
	}
}

func validateWithDatabaseResult(t *testing.T, metrics pdata.MetricSlice) {
	require.Equal(t, len(metadata.M.Names()), metrics.Len())

	for i := 0; i < metrics.Len(); i++ {
		m := metrics.At(i)
		switch m.Name() {
		case metadata.M.MysqlBufferPoolPages.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 6, dps.Len())
			bufferPoolPagesMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolPages)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				bufferPoolPagesMetrics[label] = true
			}
			require.Equal(t, map[string]bool{
				"mysql.buffer_pool_pages :data database:otel":    true,
				"mysql.buffer_pool_pages :dirty database:otel":   true,
				"mysql.buffer_pool_pages :flushed database:otel": true,
				"mysql.buffer_pool_pages :free database:otel":    true,
				"mysql.buffer_pool_pages :misc database:otel":    true,
				"mysql.buffer_pool_pages :total database:otel":   true,
			}, bufferPoolPagesMetrics)
		case metadata.M.MysqlBufferPoolOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 7, dps.Len())
			bufferPoolOperationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolOperations)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				bufferPoolOperationsMetrics[label] = true
			}
			require.Equal(t, 7, len(bufferPoolOperationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.buffer_pool_operations :read_ahead database:otel":         true,
				"mysql.buffer_pool_operations :read_ahead_evicted database:otel": true,
				"mysql.buffer_pool_operations :read_ahead_rnd database:otel":     true,
				"mysql.buffer_pool_operations :read_requests database:otel":      true,
				"mysql.buffer_pool_operations :reads database:otel":              true,
				"mysql.buffer_pool_operations :wait_free database:otel":          true,
				"mysql.buffer_pool_operations :write_requests database:otel":     true,
			}, bufferPoolOperationsMetrics)
		case metadata.M.MysqlBufferPoolSize.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 3, dps.Len())
			bufferPoolSizeMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.BufferPoolSize)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				bufferPoolSizeMetrics[label] = true
			}
			require.Equal(t, 3, len(bufferPoolSizeMetrics))
			require.Equal(t, map[string]bool{
				"mysql.buffer_pool_size :data database:otel":  true,
				"mysql.buffer_pool_size :dirty database:otel": true,
				"mysql.buffer_pool_size :size database:otel":  true,
			}, bufferPoolSizeMetrics)
		case metadata.M.MysqlCommands.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 6, dps.Len())
			commandsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Command)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				commandsMetrics[label] = true
			}
			require.Equal(t, 6, len(commandsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.commands :close database:otel":          true,
				"mysql.commands :execute database:otel":        true,
				"mysql.commands :fetch database:otel":          true,
				"mysql.commands :prepare database:otel":        true,
				"mysql.commands :reset database:otel":          true,
				"mysql.commands :send_long_data database:otel": true,
			}, commandsMetrics)
		case metadata.M.MysqlHandlers.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 18, dps.Len())
			handlersMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Handler)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				handlersMetrics[label] = true
			}
			require.Equal(t, 18, len(handlersMetrics))
			require.Equal(t, map[string]bool{
				"mysql.handlers :commit database:otel":             true,
				"mysql.handlers :delete database:otel":             true,
				"mysql.handlers :discover database:otel":           true,
				"mysql.handlers :lock database:otel":               true,
				"mysql.handlers :mrr_init database:otel":           true,
				"mysql.handlers :prepare database:otel":            true,
				"mysql.handlers :read_first database:otel":         true,
				"mysql.handlers :read_key database:otel":           true,
				"mysql.handlers :read_last database:otel":          true,
				"mysql.handlers :read_next database:otel":          true,
				"mysql.handlers :read_prev database:otel":          true,
				"mysql.handlers :read_rnd database:otel":           true,
				"mysql.handlers :read_rnd_next database:otel":      true,
				"mysql.handlers :rollback database:otel":           true,
				"mysql.handlers :savepoint database:otel":          true,
				"mysql.handlers :savepoint_rollback database:otel": true,
				"mysql.handlers :update database:otel":             true,
				"mysql.handlers :write database:otel":              true,
			}, handlersMetrics)
		case metadata.M.MysqlDoubleWrites.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			doubleWritesMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.DoubleWrites)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				doubleWritesMetrics[label] = true
			}
			require.Equal(t, 2, len(doubleWritesMetrics))
			require.Equal(t, map[string]bool{
				"mysql.double_writes :writes database:otel":  true,
				"mysql.double_writes :written database:otel": true,
			}, doubleWritesMetrics)
		case metadata.M.MysqlLogOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			logOperationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.LogOperations)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				logOperationsMetrics[label] = true
			}
			require.Equal(t, 3, len(logOperationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.log_operations :requests database:otel": true,
				"mysql.log_operations :waits database:otel":    true,
				"mysql.log_operations :writes database:otel":   true,
			}, logOperationsMetrics)
		case metadata.M.MysqlOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			operationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Operations)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				operationsMetrics[label] = true
			}
			require.Equal(t, 3, len(operationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.operations :fsyncs database:otel": true,
				"mysql.operations :reads database:otel":  true,
				"mysql.operations :writes database:otel": true,
			}, operationsMetrics)
		case metadata.M.MysqlPageOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 3, dps.Len())
			pageOperationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.PageOperations)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				pageOperationsMetrics[label] = true
			}
			require.Equal(t, 3, len(pageOperationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.page_operations :created database:otel": true,
				"mysql.page_operations :read database:otel":    true,
				"mysql.page_operations :written database:otel": true,
			}, pageOperationsMetrics)
		case metadata.M.MysqlRowLocks.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			rowLocksMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.RowLocks)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				rowLocksMetrics[label] = true
			}
			require.Equal(t, 2, len(rowLocksMetrics))
			require.Equal(t, map[string]bool{
				"mysql.row_locks :time database:otel":  true,
				"mysql.row_locks :waits database:otel": true,
			}, rowLocksMetrics)
		case metadata.M.MysqlRowOperations.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 4, dps.Len())
			rowOperationsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.RowOperations)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				rowOperationsMetrics[label] = true
			}
			require.Equal(t, 4, len(rowOperationsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.row_operations :deleted database:otel":  true,
				"mysql.row_operations :inserted database:otel": true,
				"mysql.row_operations :read database:otel":     true,
				"mysql.row_operations :updated database:otel":  true,
			}, rowOperationsMetrics)
		case metadata.M.MysqlLocks.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 2, dps.Len())
			locksMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Locks)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				locksMetrics[label] = true
			}
			require.Equal(t, 2, len(locksMetrics))
			require.Equal(t, map[string]bool{
				"mysql.locks :immediate database:otel": true,
				"mysql.locks :waited database:otel":    true,
			}, locksMetrics)
		case metadata.M.MysqlSorts.Name():
			dps := m.Sum().DataPoints()
			require.True(t, m.Sum().IsMonotonic())
			require.Equal(t, 4, dps.Len())
			sortsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Sorts)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				sortsMetrics[label] = true
			}
			require.Equal(t, 4, len(sortsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.sorts :merge_passes database:otel": true,
				"mysql.sorts :range database:otel":        true,
				"mysql.sorts :rows database:otel":         true,
				"mysql.sorts :scan database:otel":         true,
			}, sortsMetrics)
		case metadata.M.MysqlThreads.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 4, dps.Len())
			threadsMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				value_label, _ := dp.LabelsMap().Get(metadata.L.Threads)
				db_label, _ := dp.LabelsMap().Get(metadata.L.Database)
				label := fmt.Sprintf("%s :%s database:%s", m.Name(), value_label, db_label)
				threadsMetrics[label] = true
			}
			require.Equal(t, 4, len(threadsMetrics))
			require.Equal(t, map[string]bool{
				"mysql.threads :cached database:otel":    true,
				"mysql.threads :connected database:otel": true,
				"mysql.threads :created database:otel":   true,
				"mysql.threads :running database:otel":   true,
			}, threadsMetrics)
		}
	}
}

func TestMySQLStartStop(t *testing.T) {
	container := getContainer(t, containerRequest8_0)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
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
