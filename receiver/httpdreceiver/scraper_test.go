package httpdreceiver

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/observiq/opentelemetry-components/receiver/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configtls"
	"go.uber.org/zap"
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

	expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testscraper/expected_metrics.json")
	require.NoError(t, err)

	helper.ScraperTest(t, sc.scrape, expectedFileBytes)
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
