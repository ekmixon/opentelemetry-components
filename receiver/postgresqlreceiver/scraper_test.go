package postgresqlreceiver

import (
	"io/ioutil"
	"testing"

	"github.com/observiq/opentelemetry-components/receiver/helper"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestScraper(t *testing.T) {
	postgresqlMock := fakeClient{database: "otel"}
	sc := newPostgreSQLScraper(zap.NewNop(), &Config{})
	sc.client = &postgresqlMock

	expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testscraper/expected_metrics.json")
	require.NoError(t, err)

	helper.ScraperTest(t, sc.scrape, expectedFileBytes)
}
