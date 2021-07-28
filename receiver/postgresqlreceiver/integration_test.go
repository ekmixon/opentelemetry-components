// +build integration

package postgresqlreceiver

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

	"github.com/observiq/opentelemetry-components/receiver/postgresqlreceiver/internal/metadata"
)

func postgresqlContainer(t *testing.T) testcontainers.Container {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.postgresql",
		},
		ExposedPorts: []string{"5432:5432"},
		WaitingFor:   wait.ForListeningPort("5432"),
	}
	require.NoError(t, req.Validate())

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	return container
}

type PostgreSQLIntegrationSuite struct {
	suite.Suite
}

func TestPostgreSQLIntegration(t *testing.T) {
	suite.Run(t, new(PostgreSQLIntegrationSuite))
}

func (suite *PostgreSQLIntegrationSuite) TestHappyPath() {
	t := suite.T()
	container := postgresqlContainer(t)
	defer func() {
		err := container.Terminate(context.Background())
		require.NoError(t, err)
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	sc := newPostgreSQLScraper(zap.NewNop(), &Config{
		Username: "otel",
		Password: "otel",
		Database: "otel",
		Endpoint: fmt.Sprintf("%s:5432", hostname),
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
		case metadata.M.PostgresqlBlocksRead.Name():
			dps := m.IntSum().DataPoints()
			require.Equal(t, 24, dps.Len())
			metrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				dbLabel, _ := dp.LabelsMap().Get(metadata.L.Database)
				tableLabel, _ := dp.LabelsMap().Get(metadata.L.Table)
				sourceLabel, _ := dp.LabelsMap().Get(metadata.L.Source)
				label := fmt.Sprintf("%s %s %s %s", m.Name(), dbLabel, tableLabel, sourceLabel)
				metrics[label] = true
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
			dps := m.IntSum().DataPoints()
			require.Equal(t, 1, dps.Len())

			dp := dps.At(0)
			dbLabel, _ := dp.LabelsMap().Get(metadata.L.Database)
			label := fmt.Sprintf("%s %s %v", m.Name(), dbLabel, true)
			require.Equal(t, "postgresql.commits otel true", label)

		case metadata.M.PostgresqlDbSize.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())

			dp := dps.At(0)
			dbLabel, _ := dp.LabelsMap().Get(metadata.L.Database)
			label := fmt.Sprintf("%s %s %v", m.Name(), dbLabel, true)
			require.Equal(t, "postgresql.db_size otel true", label)

		case metadata.M.PostgresqlBackends.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())

			dp := dps.At(0)
			dbLabel, _ := dp.LabelsMap().Get(metadata.L.Database)
			label := fmt.Sprintf("%s %s %v", m.Name(), dbLabel, true)
			require.Equal(t, "postgresql.backends otel true", label)

		case metadata.M.PostgresqlRows.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 6, dps.Len())

			metrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				dbLabel, _ := dp.LabelsMap().Get(metadata.L.Database)
				tableLabel, _ := dp.LabelsMap().Get(metadata.L.Table)
				stateLabel, _ := dp.LabelsMap().Get(metadata.L.State)
				label := fmt.Sprintf("%s %s %s %s", m.Name(), dbLabel, tableLabel, stateLabel)
				metrics[label] = true
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
			dps := m.IntSum().DataPoints()
			require.Equal(t, 12, dps.Len())

			metrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				dbLabel, _ := dp.LabelsMap().Get(metadata.L.Database)
				tableLabel, _ := dp.LabelsMap().Get(metadata.L.Table)
				operationLabel, _ := dp.LabelsMap().Get(metadata.L.Operation)
				label := fmt.Sprintf("%s %s %s %s", m.Name(), dbLabel, tableLabel, operationLabel)
				metrics[label] = true
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
			dbLabel, _ := dp.LabelsMap().Get(metadata.L.Database)
			label := fmt.Sprintf("%s %s %v", m.Name(), dbLabel, true)
			require.Equal(t, "postgresql.rollbacks otel true", label)

		default:
			require.Nil(t, m.Name(), fmt.Sprintf("metrics %s not expected", m.Name()))
		}
	}
}

func (suite *PostgreSQLIntegrationSuite) TestStartStop() {
	t := suite.T()
	container := postgresqlContainer(t)
	defer func() {
		err := container.Terminate(context.Background())
		require.NoError(t, err)
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	sc := newPostgreSQLScraper(zap.NewNop(), &Config{
		Username: "otel",
		Password: "otel",
		Database: "otel",
		Endpoint: fmt.Sprintf("%s:5432", hostname),
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
			dps := m.IntSum().DataPoints()
			require.Equal(t, 0, dps.Len())

		case metadata.M.PostgresqlCommits.Name():
			dps := m.IntSum().DataPoints()
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
			dps := m.IntSum().DataPoints()
			require.Equal(t, 0, dps.Len())

		case metadata.M.PostgresqlRollbacks.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 0, dps.Len())

		default:
			require.Nil(t, m.Name(), fmt.Sprintf("metrics %s not expected", m.Name()))
		}
	}
}
