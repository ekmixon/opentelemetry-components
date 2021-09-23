package mongodbreceiver

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/observiq/opentelemetry-components/receiver/mongodbreceiver/internal/metadata"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestScraper(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = net.JoinHostPort("localhost", "37017")
	cfg.Username = "otel"
	cfg.Password = "otel"

	sc := newMongodbScraper(zap.NewNop(), cfg)
	sc.client = &fakeClient{}
	rms, err := sc.scrape(context.Background())
	require.Nil(t, err)
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
			require.Equal(t, 1, m.Gauge().DataPoints().Len())
			require.Equal(t, int64(1), m.Gauge().DataPoints().At(0).IntVal())
			dbName, ok := m.Gauge().DataPoints().At(0).Attributes().Get(metadata.L.DatabaseName)
			if !ok {
				require.NoError(t, fmt.Errorf("database name label missing"))
			}
			require.Equal(t, "fakedatabase", dbName.AsString())
		case "mongodb.data_size":
			require.Equal(t, 1, m.Gauge().DataPoints().Len())
			require.Equal(t, float64(3141), m.Gauge().DataPoints().At(0).DoubleVal())
			dbName, ok := m.Gauge().DataPoints().At(0).Attributes().Get(metadata.L.DatabaseName)
			if !ok {
				require.NoError(t, fmt.Errorf("database name label missing"))
			}
			require.Equal(t, "fakedatabase", dbName.AsString())
		case "mongodb.extents":
			require.Equal(t, 1, m.Gauge().DataPoints().Len())
			require.Equal(t, int64(0), m.Gauge().DataPoints().At(0).IntVal())
			dbName, ok := m.Gauge().DataPoints().At(0).Attributes().Get(metadata.L.DatabaseName)
			if !ok {
				require.NoError(t, fmt.Errorf("database name label missing"))
			}
			require.Equal(t, "fakedatabase", dbName.AsString())
		case "mongodb.index_size":
			require.Equal(t, 1, m.Gauge().DataPoints().Len())
			require.Equal(t, float64(16384), m.Gauge().DataPoints().At(0).DoubleVal())
			dbName, ok := m.Gauge().DataPoints().At(0).Attributes().Get(metadata.L.DatabaseName)
			if !ok {
				require.NoError(t, fmt.Errorf("database name label missing"))
			}
			require.Equal(t, "fakedatabase", dbName.AsString())
		case "mongodb.indexes":
			require.Equal(t, 1, m.Gauge().DataPoints().Len())
			require.Equal(t, int64(1), m.Gauge().DataPoints().At(0).IntVal())
			dbName, ok := m.Gauge().DataPoints().At(0).Attributes().Get(metadata.L.DatabaseName)
			if !ok {
				require.NoError(t, fmt.Errorf("database name label missing"))
			}
			require.Equal(t, "fakedatabase", dbName.AsString())
		case "mongodb.objects":
			require.Equal(t, 1, m.Gauge().DataPoints().Len())
			require.Equal(t, int64(2), m.Gauge().DataPoints().At(0).IntVal())
			dbName, ok := m.Gauge().DataPoints().At(0).Attributes().Get(metadata.L.DatabaseName)
			if !ok {
				require.NoError(t, fmt.Errorf("database name label missing"))
			}
			require.Equal(t, "fakedatabase", dbName.AsString())
		case "mongodb.storage_size":
			require.Equal(t, 1, m.Gauge().DataPoints().Len())
			require.Equal(t, float64(16384), m.Gauge().DataPoints().At(0).DoubleVal())
			dbName, ok := m.Gauge().DataPoints().At(0).Attributes().Get(metadata.L.DatabaseName)
			if !ok {
				require.NoError(t, fmt.Errorf("database name label missing"))
			}
			require.Equal(t, "fakedatabase", dbName.AsString())
		case "mongodb.connections":
			require.Equal(t, 1, m.Gauge().DataPoints().Len())
			require.Equal(t, int64(3), m.Gauge().DataPoints().At(0).IntVal())
			dbName, ok := m.Gauge().DataPoints().At(0).Attributes().Get(metadata.L.DatabaseName)
			if !ok {
				require.NoError(t, fmt.Errorf("database name label missing"))
			}
			require.Equal(t, "fakedatabase", dbName.AsString())
		case "mongodb.memory_usage":
			require.Equal(t, 4, m.Gauge().DataPoints().Len())
			dps := m.Gauge().DataPoints()
			for dpIndex := 0; dpIndex < dps.Len(); dpIndex++ {
				dp := dps.At(dpIndex)

				dbName, ok := dp.Attributes().Get(metadata.L.DatabaseName)
				if !ok {
					require.NoError(t, fmt.Errorf("database name label missing"))
				}
				require.Equal(t, "fakedatabase", dbName.AsString())

				dpLabel, ok := dp.Attributes().Get(metadata.L.MemoryType)
				if !ok {
					require.NoError(t, fmt.Errorf("Memory type label doesn't exist where it should"))
				}
				switch dpLabel.AsString() {
				case "resident":
					require.Equal(t, int64(79), dp.IntVal())
				case "virtual":
					require.Equal(t, int64(1089), dp.IntVal())
				case "mapped":
					require.Equal(t, int64(0), dp.IntVal())
				case "mappedWithJournal":
					require.Equal(t, int64(0), dp.IntVal())
				default:
					require.NoError(t, fmt.Errorf("Incorrect memory type"))
				}
			}
		case "mongodb.global_lock_hold_time":
			require.Equal(t, 1, m.Sum().DataPoints().Len())
			require.Equal(t, int64(58964000), m.Sum().DataPoints().At(0).IntVal())
		case "mongodb.cache_misses":
			require.Equal(t, 1, m.Sum().DataPoints().Len())
			require.Equal(t, int64(18), m.Sum().DataPoints().At(0).IntVal())
		case "mongodb.cache_hits":
			require.Equal(t, 1, m.Sum().DataPoints().Len())
			require.Equal(t, int64(197), m.Sum().DataPoints().At(0).IntVal())
		case "mongodb.operation_count":
			require.Equal(t, 6, m.Sum().DataPoints().Len())
			dps := m.Sum().DataPoints()
			for dpIndex := 0; dpIndex < dps.Len(); dpIndex++ {
				dp := dps.At(dpIndex)
				dpLabel, ok := dp.Attributes().Get(metadata.L.Operation)
				if !ok {
					require.NoError(t, fmt.Errorf("Operation label doesn't exist where it should"))
				}
				switch dpLabel.AsString() {
				case "insert":
					require.Equal(t, int64(0), dp.IntVal())
				case "query":
					require.Equal(t, int64(2), dp.IntVal())
				case "update":
					require.Equal(t, int64(0), dp.IntVal())
				case "delete":
					require.Equal(t, int64(0), dp.IntVal())
				case "getmore":
					require.Equal(t, int64(0), dp.IntVal())
				case "command":
					require.Equal(t, int64(20), dp.IntVal())
				default:
					require.NoError(t, fmt.Errorf("Incorrect operation"))
				}
			}
		default:
			t.Errorf("Incorrect name or untracked metric name %s", m.Name())
		}
	}
}
