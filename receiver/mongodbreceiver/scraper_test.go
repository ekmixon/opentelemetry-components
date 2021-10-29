package mongodbreceiver

import (
	"io/ioutil"
	"net"
	"testing"

	"github.com/observiq/opentelemetry-components/receiver/helper"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestScrape(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = net.JoinHostPort("localhost", "37017")
	cfg.Username = "otel"
	cfg.Password = "otel"

	scraper := mongodbScraper{
		logger:      zap.NewNop(),
		config:      cfg,
		buildClient: createFakeClient,
	}

	expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testscraper/expected_metrics.json")
	require.NoError(t, err)

	helper.ScraperTest(t, scraper.scrape, expectedFileBytes)
}
