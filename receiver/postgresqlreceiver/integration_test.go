package postgresqlreceiver

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.uber.org/zap"

	"github.com/observiq/opentelemetry-components/receiver/helper"
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

	expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testintegration/expected_metrics.json")
	require.NoError(t, err)

	helper.IntegrationTestHelper(t, cfg, f, expectedFileBytes, map[string]bool{})
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
