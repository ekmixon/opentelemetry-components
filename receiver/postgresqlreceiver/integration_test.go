//go:build integration
// +build integration

package postgresqlreceiver

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

	"github.com/observiq/opentelemetry-components/receiver/postgresqlreceiver/internal/metadata"
)

func TestPostgreSQLIntegration(t *testing.T) {
	container := getContainer(t, containerRequest9_6)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = net.JoinHostPort(hostname, "5432")
	cfg.Database = "otel"
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

	validateResult(t, metrics)
}

var (
	containerRequest9_6 = testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.postgresql",
		},
		ExposedPorts: []string{"5432:5432"},
		WaitingFor: wait.ForListeningPort("5432").
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
	return container
}

func validateResult(t *testing.T, metrics pdata.MetricSlice) {
	require.Equal(t, len(metadata.M.Names()), metrics.Len())

	for i := 0; i < metrics.Len(); i++ {
		m := metrics.At(i)
		switch m.Name() {
		case metadata.M.PostgresqlBlocksRead.Name():
			dps := m.Sum().DataPoints()
			require.Equal(t, 24, dps.Len())
			metrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				dbAttribute, _ := dp.Attributes().Get(metadata.L.Database)
				tableAttribute, _ := dp.Attributes().Get(metadata.L.Table)
				sourceAttribute, _ := dp.Attributes().Get(metadata.L.Source)
				attribute := fmt.Sprintf("%s %s %s %s", m.Name(), dbAttribute.AsString(), tableAttribute.AsString(), sourceAttribute.AsString())
				metrics[attribute] = true
			}
			require.Equal(t, 24, len(metrics))
			require.Equal(t, map[string]bool{
				"postgresql.blocks_read otel _global heap_read":        true,
				"postgresql.blocks_read otel _global heap_hit":         true,
				"postgresql.blocks_read otel _global idx_read":         true,
				"postgresql.blocks_read otel _global idx_hit":          true,
				"postgresql.blocks_read otel _global toast_read":       true,
				"postgresql.blocks_read otel _global toast_hit":        true,
				"postgresql.blocks_read otel _global tidx_read":        true,
				"postgresql.blocks_read otel _global tidx_hit":         true,
				"postgresql.blocks_read otel public.table1 heap_read":  true,
				"postgresql.blocks_read otel public.table1 heap_hit":   true,
				"postgresql.blocks_read otel public.table1 idx_read":   true,
				"postgresql.blocks_read otel public.table1 idx_hit":    true,
				"postgresql.blocks_read otel public.table1 toast_read": true,
				"postgresql.blocks_read otel public.table1 toast_hit":  true,
				"postgresql.blocks_read otel public.table1 tidx_read":  true,
				"postgresql.blocks_read otel public.table1 tidx_hit":   true,
				"postgresql.blocks_read otel public.table2 heap_read":  true,
				"postgresql.blocks_read otel public.table2 heap_hit":   true,
				"postgresql.blocks_read otel public.table2 idx_read":   true,
				"postgresql.blocks_read otel public.table2 idx_hit":    true,
				"postgresql.blocks_read otel public.table2 toast_read": true,
				"postgresql.blocks_read otel public.table2 toast_hit":  true,
				"postgresql.blocks_read otel public.table2 tidx_read":  true,
				"postgresql.blocks_read otel public.table2 tidx_hit":   true,
			}, metrics)

		case metadata.M.PostgresqlCommits.Name():
			dps := m.Sum().DataPoints()
			require.Equal(t, 1, dps.Len())

			dp := dps.At(0)
			dbAttribute, _ := dp.Attributes().Get(metadata.L.Database)
			attribute := fmt.Sprintf("%s %s %v", m.Name(), dbAttribute.AsString(), true)
			require.Equal(t, "postgresql.commits otel true", attribute)

		case metadata.M.PostgresqlDbSize.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())

			dp := dps.At(0)
			dbAttribute, _ := dp.Attributes().Get(metadata.L.Database)
			attribute := fmt.Sprintf("%s %s %v", m.Name(), dbAttribute.AsString(), true)
			require.Equal(t, "postgresql.db_size otel true", attribute)

		case metadata.M.PostgresqlBackends.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())

			dp := dps.At(0)
			dbAttribute, _ := dp.Attributes().Get(metadata.L.Database)
			attribute := fmt.Sprintf("%s %s %v", m.Name(), dbAttribute.AsString(), true)
			require.Equal(t, "postgresql.backends otel true", attribute)

		case metadata.M.PostgresqlRows.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 6, dps.Len())

			metrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				dbAttribute, _ := dp.Attributes().Get(metadata.L.Database)
				tableAttribute, _ := dp.Attributes().Get(metadata.L.Table)
				stateAttribute, _ := dp.Attributes().Get(metadata.L.State)
				attribute := fmt.Sprintf("%s %s %s %s", m.Name(), dbAttribute.AsString(), tableAttribute.AsString(), stateAttribute.AsString())
				metrics[attribute] = true
			}
			require.Equal(t, 6, len(metrics))
			require.Equal(t, map[string]bool{
				"postgresql.rows otel _global live":       true,
				"postgresql.rows otel _global dead":       true,
				"postgresql.rows otel public.table1 live": true,
				"postgresql.rows otel public.table1 dead": true,
				"postgresql.rows otel public.table2 live": true,
				"postgresql.rows otel public.table2 dead": true,
			}, metrics)

		case metadata.M.PostgresqlOperations.Name():
			dps := m.Sum().DataPoints()
			require.Equal(t, 12, dps.Len())

			metrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				dbAttribute, _ := dp.Attributes().Get(metadata.L.Database)
				tableAttribute, _ := dp.Attributes().Get(metadata.L.Table)
				operationAttribute, _ := dp.Attributes().Get(metadata.L.Operation)
				attribute := fmt.Sprintf("%s %s %s %s", m.Name(), dbAttribute.AsString(), tableAttribute.AsString(), operationAttribute.AsString())
				metrics[attribute] = true
			}
			require.Equal(t, 12, len(metrics))
			require.Equal(t, map[string]bool{
				"postgresql.operations otel _global seq":                 true,
				"postgresql.operations otel _global seq_tup_read":        true,
				"postgresql.operations otel _global idx":                 true,
				"postgresql.operations otel _global idx_tup_fetch":       true,
				"postgresql.operations otel public.table1 seq":           true,
				"postgresql.operations otel public.table1 seq_tup_read":  true,
				"postgresql.operations otel public.table1 idx":           true,
				"postgresql.operations otel public.table1 idx_tup_fetch": true,
				"postgresql.operations otel public.table2 seq":           true,
				"postgresql.operations otel public.table2 seq_tup_read":  true,
				"postgresql.operations otel public.table2 idx":           true,
				"postgresql.operations otel public.table2 idx_tup_fetch": true,
			}, metrics)

		case metadata.M.PostgresqlRollbacks.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())

			dp := dps.At(0)
			dbAttribute, _ := dp.Attributes().Get(metadata.L.Database)
			attribute := fmt.Sprintf("%s %s %v", m.Name(), dbAttribute.AsString(), true)
			require.Equal(t, "postgresql.rollbacks otel true", attribute)

		default:
			require.Nil(t, m.Name(), fmt.Sprintf("metrics %s not expected", m.Name()))
		}
	}
}

func TestPostgreSQLStartStop(t *testing.T) {
	container := getContainer(t, containerRequest9_6)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	sc := newPostgreSQLScraper(zap.NewNop(), &Config{
		Endpoint: fmt.Sprintf("%s:5432", hostname),
		Database: "otel",
		Username: "otel",
		Password: "otel",
	})

	// scraper is connected
	err = sc.start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	// scraper is closed
	err = sc.shutdown(context.Background())
	require.NoError(t, err)

	// scraper scapes without a db connection and collect 0 metrics
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
		case metadata.M.PostgresqlBlocksRead.Name():
			dps := m.Sum().DataPoints()
			require.Equal(t, 0, dps.Len())

		case metadata.M.PostgresqlCommits.Name():
			dps := m.Sum().DataPoints()
			require.Equal(t, 0, dps.Len())

		case metadata.M.PostgresqlDbSize.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 0, dps.Len())

		case metadata.M.PostgresqlBackends.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 0, dps.Len())

		case metadata.M.PostgresqlRows.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 0, dps.Len())

		case metadata.M.PostgresqlOperations.Name():
			dps := m.Sum().DataPoints()
			require.Equal(t, 0, dps.Len())

		case metadata.M.PostgresqlRollbacks.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 0, dps.Len())

		default:
			require.Nil(t, m.Name(), fmt.Sprintf("metrics %s not expected", m.Name()))
		}
	}
}
