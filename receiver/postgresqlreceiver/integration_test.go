//go:build integration
// +build integration

package postgresqlreceiver

import (
	"context"
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/model/pdata"

	"github.com/observiq/opentelemetry-components/receiver/postgresqlreceiver/internal/metadata"
)

type expectedDatabase struct {
	name   string
	tables []string
}

func TestPostgreSQLIntegrationSingle(t *testing.T) {
	container := getContainer(t, containerRequest9_6)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Host = hostname
	cfg.Port = 15432
	cfg.Databases = []string{"otel"}
	cfg.Username = "otel"
	cfg.Password = "otel"
	cfg.SSLConfig.SSLMode = "disable"

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

	validateResult(t, metrics, []expectedDatabase{
		{
			name:   "otel",
			tables: []string{"public.table1", "public.table2"},
		},
	})
}

func TestPostgreSQLIntegrationMultiple(t *testing.T) {
	container := getContainer(t, containerRequest9_6)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Host = hostname
	cfg.Port = 15432
	cfg.Databases = []string{"otel", "otel2"}
	cfg.Username = "otel"
	cfg.Password = "otel"
	cfg.SSLConfig.SSLMode = "disable"

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

	validateResult(t, metrics, []expectedDatabase{
		{
			name:   "otel",
			tables: []string{"public.table1", "public.table2"},
		},
		{
			name:   "otel2",
			tables: []string{"public.test1", "public.test2"},
		},
	})
}

func TestPostgreSQLIntegrationAll(t *testing.T) {
	container := getContainer(t, containerRequest9_6)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Host = hostname
	cfg.Port = 15432
	cfg.Databases = []string{}
	cfg.Username = "otel"
	cfg.Password = "otel"
	cfg.SSLConfig.SSLMode = "disable"

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

	validateResult(t, metrics, []expectedDatabase{
		{
			name:   "otel",
			tables: []string{"public.table1", "public.table2"},
		},
		{
			name:   "otel2",
			tables: []string{"public.test1", "public.test2"},
		},
		{
			name:   "postgres",
			tables: []string{},
		},
	})
}

var (
	containerRequest9_6 = testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.postgresql",
		},
		ExposedPorts: []string{"15432:5432"},
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

func enumerateExpectedMetrics(databases []expectedDatabase, perTableMetric bool, metric string, labels []string) map[string]bool {
	expected := map[string]bool{}
	for _, db := range databases {
		if perTableMetric {
			for _, table := range db.tables {
				if len(labels) > 0 {
					for _, label := range labels {
						expected[fmt.Sprintf("%s %s %s %s", metric, db.name, table, label)] = true
					}
				} else {
					expected[fmt.Sprintf("%s %s %s", metric, db.name, table)] = true
				}
			}
		} else {
			if len(labels) > 0 {
				for _, label := range labels {
					expected[fmt.Sprintf("%s %s %s", metric, db.name, label)] = true
				}
			} else {
				expected[fmt.Sprintf("%s %s", metric, db.name)] = true
			}
		}
	}
	return expected
}

func enumerateActualMetrics(m pdata.Metric, dps pdata.NumberDataPointSlice, labelKey string) map[string]bool {
	actual := map[string]bool{}
	for j := 0; j < dps.Len(); j++ {
		dp := dps.At(j)
		dbAttribute, _ := dp.Attributes().Get(metadata.L.Database)
		if tableAttribute, ok := dp.Attributes().Get(metadata.L.Table); ok {
			if labelKey != "" {
				additionalLabelAttribute, _ := dp.Attributes().Get(labelKey)
				actual[fmt.Sprintf("%s %s %s %s", m.Name(), dbAttribute.AsString(), tableAttribute.AsString(), additionalLabelAttribute.AsString())] = true
			} else {
				actual[fmt.Sprintf("%s %s %s", m.Name(), dbAttribute.AsString(), tableAttribute.AsString())] = true
			}
		} else {
			if labelKey != "" {
				additionalLabelAttribute, _ := dp.Attributes().Get(labelKey)
				actual[fmt.Sprintf("%s %s %s", m.Name(), dbAttribute.AsString(), additionalLabelAttribute.AsString())] = true
			} else {
				actual[fmt.Sprintf("%s %s", m.Name(), dbAttribute.AsString())] = true
			}
		}
	}
	return actual
}

func validateResult(t *testing.T, metrics pdata.MetricSlice, databases []expectedDatabase) {
	require.Equal(t, len(metadata.M.Names()), metrics.Len())

	for i := 0; i < metrics.Len(); i++ {
		m := metrics.At(i)
		switch m.Name() {
		case metadata.M.PostgresqlBlocksRead.Name():
			actual := enumerateActualMetrics(m, m.Sum().DataPoints(), metadata.L.Source)
			expected := enumerateExpectedMetrics(
				databases,
				true,
				"postgresql.blocks_read",
				[]string{"heap_read", "heap_hit", "idx_read", "idx_hit", "toast_read", "toast_hit", "tidx_read", "tidx_hit"},
			)
			require.Equal(t, len(expected), len(actual))
			require.Equal(t, expected, actual)

		case metadata.M.PostgresqlCommits.Name():
			actual := enumerateActualMetrics(m, m.Sum().DataPoints(), "")
			expected := enumerateExpectedMetrics(
				databases,
				false,
				"postgresql.commits",
				[]string{},
			)

			require.Equal(t, len(expected), len(actual))
			require.Equal(t, expected, actual)

		case metadata.M.PostgresqlDbSize.Name():
			actual := enumerateActualMetrics(m, m.Gauge().DataPoints(), "")
			expected := enumerateExpectedMetrics(
				databases,
				false,
				"postgresql.db_size",
				[]string{},
			)

			require.Equal(t, len(expected), len(actual))
			require.Equal(t, expected, actual)

		case metadata.M.PostgresqlBackends.Name():
			actual := enumerateActualMetrics(m, m.Gauge().DataPoints(), "")
			// Only expect current connections for the root DB as our connections are all closed
			expected := enumerateExpectedMetrics(
				[]expectedDatabase{databases[0]},
				false,
				"postgresql.backends",
				[]string{},
			)

			require.Equal(t, len(expected), len(actual))
			require.Equal(t, expected, actual)

		case metadata.M.PostgresqlRows.Name():
			actual := enumerateActualMetrics(m, m.Gauge().DataPoints(), metadata.L.State)
			expected := enumerateExpectedMetrics(
				databases,
				true,
				"postgresql.rows",
				[]string{"live", "dead"},
			)
			require.Equal(t, len(expected), len(actual))
			require.Equal(t, expected, actual)

		case metadata.M.PostgresqlOperations.Name():
			actual := enumerateActualMetrics(m, m.Sum().DataPoints(), metadata.L.Operation)
			expected := enumerateExpectedMetrics(
				databases,
				true,
				"postgresql.operations",
				[]string{"del", "hot_upd", "ins", "upd"},
			)
			require.Equal(t, len(expected), len(actual))
			require.Equal(t, expected, actual)

		case metadata.M.PostgresqlRollbacks.Name():
			actual := enumerateActualMetrics(m, m.Sum().DataPoints(), "")
			expected := enumerateExpectedMetrics(
				databases,
				false,
				"postgresql.rollbacks",
				[]string{},
			)

			require.Equal(t, len(expected), len(actual))
			require.Equal(t, expected, actual)

		default:
			require.Nil(t, m.Name(), fmt.Sprintf("metric %s not expected", m.Name()))
		}
	}
}
