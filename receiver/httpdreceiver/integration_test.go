// +build integration

package httpdreceiver

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/zap"

	"github.com/observiq/opentelemetry-components/receiver/httpdreceiver/internal/metadata"
)

type HttpdIntegrationSuite struct {
	suite.Suite
}

func TestHttpdIntegration(t *testing.T) {
	suite.Run(t, new(HttpdIntegrationSuite))
}

type waitStrategy struct{}

func (ws waitStrategy) WaitUntilReady(ctx context.Context, st wait.StrategyTarget) error {
	if err := wait.ForListeningPort("80").WaitUntilReady(ctx, st); err != nil {
		return err
	}

	hostname, err := st.Host(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			return fmt.Errorf("server startup problem")
		case <-time.After(100 * time.Millisecond):
			resp, err := http.Get(fmt.Sprintf("http://%s:8080/server-status?auto", hostname))
			if err != nil {
				continue
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				continue
			}

			if resp.Body.Close() != nil {
				continue
			}

			// The server needs a moment to generate some stats
			if strings.Contains(string(body), "ReqPerSec") {
				return nil
			}
		}
	}

	return nil
}

func httpdContainer(t *testing.T) testcontainers.Container {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.httpd",
		},
		ExposedPorts: []string{"8080:80"},
		WaitingFor:   waitStrategy{},
	}

	require.NoError(t, req.Validate())

	httpd, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	return httpd
}

func (suite *HttpdIntegrationSuite) TestHttpdScraperHappyPath() {
	t := suite.T()
	httpd := httpdContainer(t)
	defer func() {
		require.NoError(t, httpd.Terminate(context.Background()))
	}()
	hostname, err := httpd.Host(context.Background())
	require.NoError(t, err)

	cfg := &Config{
		ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
			CollectionInterval: 100 * time.Millisecond,
		},
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: fmt.Sprintf("http://%s:8080/server-status?auto", hostname),
		},
	}

	sc := newHttpdScraper(zap.NewNop(), cfg)
	err = sc.start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)
	rms, err := sc.scrape(context.Background())
	require.Nil(t, err)

	require.Equal(t, 1, rms.Len())
	rm := rms.At(0)

	ilms := rm.InstrumentationLibraryMetrics()
	require.EqualValues(t, 1, ilms.Len())

	ilm := ilms.At(0)
	ms := ilm.Metrics()

	require.EqualValues(t, 7, ms.Len())

	for i := 0; i < ms.Len(); i++ {
		m := ms.At(i)

		switch m.Name() {
		case metadata.M.HttpdUptime.Name():
			require.Equal(t, 1, m.IntSum().DataPoints().Len())
		case metadata.M.HttpdCurrentConnections.Name():
			require.Equal(t, 1, m.IntGauge().DataPoints().Len())
		case metadata.M.HttpdWorkers.Name():
			require.Equal(t, 2, m.IntGauge().DataPoints().Len())
			dps := m.IntGauge().DataPoints()
			present := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get("state")
				switch state {
				case metadata.LabelWorkersState.Busy:
					present[state] = true
				case metadata.LabelWorkersState.Idle:
					present[state] = true
				default:
					require.Nil(t, state, fmt.Sprintf("connections with state %s not expected", state))
				}
			}
			require.Equal(t, 2, len(present))
		case metadata.M.HttpdRequests.Name():
			require.Equal(t, 1, m.Gauge().DataPoints().Len())
		case metadata.M.HttpdBytes.Name():
			require.Equal(t, 1, m.Gauge().DataPoints().Len())
		case metadata.M.HttpdTraffic.Name():
			require.Equal(t, 1, m.IntSum().DataPoints().Len())
			require.True(t, m.IntSum().IsMonotonic())
		case metadata.M.HttpdScoreboard.Name():
			dps := m.IntGauge().DataPoints()
			require.Equal(t, 11, dps.Len())
			present := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get("state")
				switch state {
				case metadata.LabelScoreboardState.Waiting:
					present[state] = true
				case metadata.LabelScoreboardState.Starting:
					present[state] = true
				case metadata.LabelScoreboardState.Reading:
					present[state] = true
				case metadata.LabelScoreboardState.Sending:
					present[state] = true
				case metadata.LabelScoreboardState.Keepalive:
					present[state] = true
				case metadata.LabelScoreboardState.Dnslookup:
					present[state] = true
				case metadata.LabelScoreboardState.Closing:
					present[state] = true
				case metadata.LabelScoreboardState.Logging:
					present[state] = true
				case metadata.LabelScoreboardState.Finishing:
					present[state] = true
				case metadata.LabelScoreboardState.IdleCleanup:
					present[state] = true
				case metadata.LabelScoreboardState.Open:
					present[state] = true
				default:
					require.Nil(t, state, fmt.Sprintf("connections with state %s not expected", state))
				}
			}
			require.Equal(t, 11, len(present))
		default:
			require.Nil(t, m.Name(), fmt.Sprintf("metrics %s not expected", m.Name()))
		}

	}

}
