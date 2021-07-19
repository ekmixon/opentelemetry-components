package mongodbreceiver

import (
	"context"
	"net"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.uber.org/zap"
)

func mongodbContainer(t *testing.T) testcontainers.Container {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.mongodb",
		},
		ExposedPorts: []string{"37017:27017"},
		WaitingFor:   wait.ForListeningPort("27017"),
	}

	require.NoError(t, req.Validate())

	mongodb, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	return mongodb
}

func TestScraper(t *testing.T) {
	container := mongodbContainer(t)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = net.JoinHostPort(hostname, "37017")

	user := "otel"
	pass := "otel"
	cfg.User = &user
	cfg.Password = &pass

	sc := newMongodbScraper(zap.NewNop(), cfg)

	err = sc.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)
	metrics, err := sc.Scrape(context.Background(), cfg.ID())
	require.Nil(t, err)
	rms := metrics.ResourceMetrics()
	require.Equal(t, 1, rms.Len())
	rm := rms.At(0)

	ilms := rm.InstrumentationLibraryMetrics()
	require.Equal(t, 1, ilms.Len())

	ilm := ilms.At(0)
	ms := ilm.Metrics()

	require.Equal(t, 13, ms.Len())

	for i := 0; i < ms.Len(); i++ {
		m := ms.At(i)
		switch m.Name() {
		case "mongodb.collections":
			require.Equal(t, 3, m.IntGauge().DataPoints().Len())
		case "mongodb.data_size":
			require.Equal(t, 3, m.Gauge().DataPoints().Len())
		case "mongodb.extents":
			require.Equal(t, 3, m.IntGauge().DataPoints().Len())
		case "mongodb.index_size":
			require.Equal(t, 3, m.Gauge().DataPoints().Len())
		case "mongodb.indexes":
			require.Equal(t, 3, m.IntGauge().DataPoints().Len())
		case "mongodb.objects":
			require.Equal(t, 3, m.IntGauge().DataPoints().Len())
		case "mongodb.storage_size":
			require.Equal(t, 3, m.Gauge().DataPoints().Len())
		case "mongodb.connections":
			require.Equal(t, 3, m.IntGauge().DataPoints().Len())
		case "mongodb.memory_usage":
			require.Equal(t, 12, m.IntGauge().DataPoints().Len())
		case "mongodb.global_lock_hold_time":
			require.Equal(t, 1, m.IntSum().DataPoints().Len())
		case "mongodb.cache_misses":
			require.Equal(t, 1, m.IntSum().DataPoints().Len())
		case "mongodb.cache_hits":
			require.Equal(t, 1, m.IntSum().DataPoints().Len())
		case "mongodb.operation_count":
			require.Equal(t, 6, m.IntSum().DataPoints().Len())
		default:
			t.Errorf("Incorrect name or untracked metric name %s", m.Name())
		}
	}
}
