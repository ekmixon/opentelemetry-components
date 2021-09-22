package httpdreceiver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configtls"
	"go.uber.org/zap"

	"github.com/observiq/opentelemetry-components/receiver/httpdreceiver/internal/metadata"
)

func TestScraper(t *testing.T) {
	httpdMock := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.String() == "/server-status?auto" {
			rw.WriteHeader(200)
			_, err := rw.Write([]byte(`ServerUptimeSeconds: 410
Total Accesses: 14169
Total kBytes: 20910
BusyWorkers: 13
IdleWorkers: 227
ConnsTotal: 110
Scoreboard: S_DD_L_GGG_____W__IIII_C________________W__________________________________.........................____WR______W____W________________________C______________________________________W_W____W______________R_________R________C_________WK_W________K_____W__C__________W___R______.............................................................................................................................
`))
			require.NoError(t, err)
			return
		}
		rw.WriteHeader(404)
	}))
	sc := newHttpdScraper(zap.NewNop(), &Config{
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: httpdMock.URL,
		},
	})

	err := sc.start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)
	assert.NotNil(t, sc.httpClient)

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
		case metadata.M.HttpdUptime.Name():
			dps := m.Sum().DataPoints()
			serverName, _ := dps.At(0).LabelsMap().Get(metadata.L.ServerName)
			require.EqualValues(t, "127.0.0.1", serverName)
			require.Equal(t, 1, m.Sum().DataPoints().Len())
			require.True(t, m.Sum().IsMonotonic())
			require.EqualValues(t, 410, dps.At(0).IntVal())
		case metadata.M.HttpdCurrentConnections.Name():
			dps := m.Gauge().DataPoints()
			serverName, _ := dps.At(0).LabelsMap().Get(metadata.L.ServerName)
			require.EqualValues(t, "127.0.0.1", serverName)
			require.Equal(t, 1, dps.Len())
			require.EqualValues(t, 110, dps.At(0).IntVal())
		case metadata.M.HttpdWorkers.Name():
			dps := m.Gauge().DataPoints()
			serverName, _ := dps.At(0).LabelsMap().Get(metadata.L.ServerName)
			require.EqualValues(t, "127.0.0.1", serverName)
			require.Equal(t, 2, m.Gauge().DataPoints().Len())

			workerMetrics := map[string]int{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.WorkersState)
				serverName, _ := dp.LabelsMap().Get(metadata.L.ServerName)
				label := fmt.Sprintf("%s serverName:%s state:%s", m.Name(), serverName, state)
				workerMetrics[label] = int(dp.IntVal())
			}

			require.Equal(t, 2, len(workerMetrics))
			require.Equal(t, map[string]int{
				"httpd.workers serverName:127.0.0.1 state:busy": 13,
				"httpd.workers serverName:127.0.0.1 state:idle": 227,
			}, workerMetrics)
		case metadata.M.HttpdRequests.Name():
			dps := m.Sum().DataPoints()
			serverName, _ := dps.At(0).LabelsMap().Get(metadata.L.ServerName)
			require.EqualValues(t, "127.0.0.1", serverName)
			require.Equal(t, 1, m.Sum().DataPoints().Len())
			require.True(t, m.Sum().IsMonotonic())
			require.EqualValues(t, 14169, dps.At(0).IntVal())
		case metadata.M.HttpdTraffic.Name():
			dps := m.Sum().DataPoints()
			serverName, _ := dps.At(0).LabelsMap().Get(metadata.L.ServerName)
			require.EqualValues(t, "127.0.0.1", serverName)
			require.Equal(t, 1, m.Sum().DataPoints().Len())
			require.True(t, m.Sum().IsMonotonic())
			require.EqualValues(t, 21411840, dps.At(0).IntVal())
		case metadata.M.HttpdScoreboard.Name():
			dps := m.Gauge().DataPoints()
			serverName, _ := dps.At(0).LabelsMap().Get(metadata.L.ServerName)
			require.EqualValues(t, "127.0.0.1", serverName)
			require.Equal(t, 11, dps.Len())
			scoreboardMetrics := map[string]int{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.ScoreboardState)
				serverName, _ := dp.LabelsMap().Get(metadata.L.ServerName)
				label := fmt.Sprintf("%s serverName:%s state:%s", m.Name(), serverName, state)
				scoreboardMetrics[label] = int(dp.IntVal())
			}
			require.Equal(t, 11, len(scoreboardMetrics))
			require.Equal(t, map[string]int{
				"httpd.scoreboard serverName:127.0.0.1 state:open":         150,
				"httpd.scoreboard serverName:127.0.0.1 state:waiting":      217,
				"httpd.scoreboard serverName:127.0.0.1 state:starting":     1,
				"httpd.scoreboard serverName:127.0.0.1 state:reading":      4,
				"httpd.scoreboard serverName:127.0.0.1 state:sending":      12,
				"httpd.scoreboard serverName:127.0.0.1 state:keepalive":    2,
				"httpd.scoreboard serverName:127.0.0.1 state:dnslookup":    2,
				"httpd.scoreboard serverName:127.0.0.1 state:closing":      4,
				"httpd.scoreboard serverName:127.0.0.1 state:logging":      1,
				"httpd.scoreboard serverName:127.0.0.1 state:finishing":    3,
				"httpd.scoreboard serverName:127.0.0.1 state:idle_cleanup": 4,
			}, scoreboardMetrics)

		default:
			require.Nil(t, m.Name(), fmt.Sprintf("metrics %s not expected", m.Name()))
		}
	}
}

func TestScraperFailedStart(t *testing.T) {
	sc := newHttpdScraper(zap.NewNop(), &Config{
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: "localhost:8080",
			TLSSetting: configtls.TLSClientSetting{
				TLSSetting: configtls.TLSSetting{
					CAFile: "/non/existent",
				},
			},
		},
	})
	err := sc.start(context.Background(), componenttest.NewNopHost())
	require.Error(t, err)
}

func TestParseScoreboard(t *testing.T) {

	t.Run("test freq count", func(t *testing.T) {
		scoreboard := `S_DD_L_GGG_____W__IIII_C________________W__________________________________.........................____WR______W____W________________________C______________________________________W_W____W______________R_________R________C_________WK_W________K_____W__C__________W___R______.............................................................................................................................`
		results := parseScoreboard(scoreboard)

		require.EqualValues(t, int64(150), results["open"])
		require.EqualValues(t, int64(217), results["waiting"])
		require.EqualValues(t, int64(1), results["starting"])
		require.EqualValues(t, int64(4), results["reading"])
		require.EqualValues(t, int64(12), results["sending"])
		require.EqualValues(t, int64(2), results["keepalive"])
		require.EqualValues(t, int64(2), results["dnslookup"])
		require.EqualValues(t, int64(4), results["closing"])
		require.EqualValues(t, int64(1), results["logging"])
		require.EqualValues(t, int64(3), results["finishing"])
		require.EqualValues(t, int64(4), results["idle_cleanup"])
	})

	t.Run("test empty defaults", func(t *testing.T) {
		emptyString := ""
		results := parseScoreboard(emptyString)

		require.EqualValues(t, int64(0), results["open"])
		require.EqualValues(t, int64(0), results["waiting"])
		require.EqualValues(t, int64(0), results["starting"])
		require.EqualValues(t, int64(0), results["reading"])
		require.EqualValues(t, int64(0), results["sending"])
		require.EqualValues(t, int64(0), results["keepalive"])
		require.EqualValues(t, int64(0), results["dnslookup"])
		require.EqualValues(t, int64(0), results["closing"])
		require.EqualValues(t, int64(0), results["logging"])
		require.EqualValues(t, int64(0), results["finishing"])
		require.EqualValues(t, int64(0), results["idle_cleanup"])
	})
}

func TestParseStats(t *testing.T) {
	t.Run("with empty value", func(t *testing.T) {
		emptyString := ""
		require.EqualValues(t, map[string]string{}, parseStats(emptyString))
	})
	t.Run("with multi colons", func(t *testing.T) {
		got := "CurrentTime: Thursday, 17-Jun-2021 14:06:32 UTC"
		want := map[string]string{
			"CurrentTime": "Thursday, 17-Jun-2021 14:06:32 UTC",
		}
		require.EqualValues(t, want, parseStats(got))
	})
	t.Run("with header/footer", func(t *testing.T) {
		got := `localhost
ReqPerSec: 719.771
IdleWorkers: 227
ConnsTotal: 110
		`
		want := map[string]string{
			"ReqPerSec":   "719.771",
			"IdleWorkers": "227",
			"ConnsTotal":  "110",
		}
		require.EqualValues(t, want, parseStats(got))
	})
}

func TestScraperError(t *testing.T) {
	t.Run("no client", func(t *testing.T) {
		sc := newHttpdScraper(zap.NewNop(), &Config{})
		sc.httpClient = nil

		_, err := sc.scrape(context.Background())
		require.Error(t, err)
		require.EqualValues(t, errors.New("failed to connect to http client"), err)
	})
}
