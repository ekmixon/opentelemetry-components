package postgresqlreceiver

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/observiq/opentelemetry-components/receiver/postgresqlreceiver/internal/metadata"
)

func TestScraper(t *testing.T) {
	postgresqlMock := fakeClient{database: "otel"}
	sc := newPostgreSQLScraper(zap.NewNop(), &Config{})
	sc.client = &postgresqlMock

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
			metrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				dbLabel, _ := dp.LabelsMap().Get(metadata.L.Database)
				tableLabel, _ := dp.LabelsMap().Get(metadata.L.Table)
				sourceLabel, _ := dp.LabelsMap().Get(metadata.L.Source)
				label := fmt.Sprintf("%s %s %s %s", m.Name(), dbLabel, tableLabel, sourceLabel)
				metrics[label] = dp.Value()
			}
			require.Equal(t, 24, len(metrics))
			require.Equal(t, map[string]int64{
				"postgresql.blocks_read otel _global heap_read":        11,
				"postgresql.blocks_read otel _global heap_hit":         12,
				"postgresql.blocks_read otel _global idx_read":         13,
				"postgresql.blocks_read otel _global idx_hit":          14,
				"postgresql.blocks_read otel _global toast_read":       15,
				"postgresql.blocks_read otel _global toast_hit":        16,
				"postgresql.blocks_read otel _global tidx_read":        17,
				"postgresql.blocks_read otel _global tidx_hit":         18,
				"postgresql.blocks_read otel public.table1 heap_read":  19,
				"postgresql.blocks_read otel public.table1 heap_hit":   20,
				"postgresql.blocks_read otel public.table1 idx_read":   21,
				"postgresql.blocks_read otel public.table1 idx_hit":    22,
				"postgresql.blocks_read otel public.table1 toast_read": 23,
				"postgresql.blocks_read otel public.table1 toast_hit":  24,
				"postgresql.blocks_read otel public.table1 tidx_read":  25,
				"postgresql.blocks_read otel public.table1 tidx_hit":   26,
				"postgresql.blocks_read otel public.table2 heap_read":  27,
				"postgresql.blocks_read otel public.table2 heap_hit":   28,
				"postgresql.blocks_read otel public.table2 idx_read":   29,
				"postgresql.blocks_read otel public.table2 idx_hit":    30,
				"postgresql.blocks_read otel public.table2 toast_read": 31,
				"postgresql.blocks_read otel public.table2 toast_hit":  32,
				"postgresql.blocks_read otel public.table2 tidx_read":  33,
				"postgresql.blocks_read otel public.table2 tidx_hit":   34,
			}, metrics)

		case metadata.M.PostgresqlCommits.Name():
			dps := m.IntSum().DataPoints()
			require.Equal(t, 1, dps.Len())

			dp := dps.At(0)
			dbLabel, _ := dp.LabelsMap().Get(metadata.L.Database)
			label := fmt.Sprintf("%s %s %v", m.Name(), dbLabel, dp.Value())
			require.Equal(t, "postgresql.commits otel 1", label)

		case metadata.M.PostgresqlDbSize.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())

			dp := dps.At(0)
			dbLabel, _ := dp.LabelsMap().Get(metadata.L.Database)
			label := fmt.Sprintf("%s %s %v", m.Name(), dbLabel, dp.Value())
			require.Equal(t, "postgresql.db_size otel 4", label)

		case metadata.M.PostgresqlBackends.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())

			dp := dps.At(0)
			dbLabel, _ := dp.LabelsMap().Get(metadata.L.Database)
			label := fmt.Sprintf("%s %s %v", m.Name(), dbLabel, dp.Value())
			require.Equal(t, "postgresql.backends otel 3", label)

		case metadata.M.PostgresqlRows.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 6, dps.Len())

			metrics := map[string]float64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				dbLabel, _ := dp.LabelsMap().Get(metadata.L.Database)
				tableLabel, _ := dp.LabelsMap().Get(metadata.L.Table)
				stateLabel, _ := dp.LabelsMap().Get(metadata.L.State)
				label := fmt.Sprintf("%s %s %s %s", m.Name(), dbLabel, tableLabel, stateLabel)
				metrics[label] = dp.Value()
			}
			require.Equal(t, 6, len(metrics))
			require.Equal(t, map[string]float64{
				"postgresql.rows otel _global live":       5,
				"postgresql.rows otel _global dead":       6,
				"postgresql.rows otel public.table1 live": 7,
				"postgresql.rows otel public.table1 dead": 8,
				"postgresql.rows otel public.table2 live": 9,
				"postgresql.rows otel public.table2 dead": 10,
			}, metrics)

		case metadata.M.PostgresqlOperations.Name():
			dps := m.IntSum().DataPoints()
			require.Equal(t, 12, dps.Len())

			metrics := map[string]int64{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				dbLabel, _ := dp.LabelsMap().Get(metadata.L.Database)
				tableLabel, _ := dp.LabelsMap().Get(metadata.L.Table)
				operationLabel, _ := dp.LabelsMap().Get(metadata.L.Operation)
				label := fmt.Sprintf("%s %s %s %s", m.Name(), dbLabel, tableLabel, operationLabel)
				metrics[label] = dp.Value()
			}
			require.Equal(t, 12, len(metrics))
			require.Equal(t, map[string]int64{
				"postgresql.operations otel _global seq":                 35,
				"postgresql.operations otel _global seq_tup_read":        36,
				"postgresql.operations otel _global idx":                 37,
				"postgresql.operations otel _global idx_tup_fetch":       38,
				"postgresql.operations otel public.table1 seq":           39,
				"postgresql.operations otel public.table1 seq_tup_read":  40,
				"postgresql.operations otel public.table1 idx":           41,
				"postgresql.operations otel public.table1 idx_tup_fetch": 42,
				"postgresql.operations otel public.table2 seq":           43,
				"postgresql.operations otel public.table2 seq_tup_read":  44,
				"postgresql.operations otel public.table2 idx":           45,
				"postgresql.operations otel public.table2 idx_tup_fetch": 46,
			}, metrics)

		case metadata.M.PostgresqlRollbacks.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())

			dp := dps.At(0)
			dbLabel, _ := dp.LabelsMap().Get(metadata.L.Database)
			label := fmt.Sprintf("%s %s %v", m.Name(), dbLabel, dp.Value())
			require.Equal(t, "postgresql.rollbacks otel 2", label)

		default:
			require.Nil(t, m.Name(), fmt.Sprintf("metrics %s not expected", m.Name()))
		}
	}
}
