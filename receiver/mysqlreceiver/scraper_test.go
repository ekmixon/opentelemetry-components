package mysqlreceiver

import (
	"io/ioutil"
	"testing"

	"github.com/observiq/opentelemetry-components/receiver/helper"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestScrape(t *testing.T) {
	mysqlMock := fakeClient{}
	sc := newMySQLScraper(zap.NewNop(), &Config{
		Username: "otel",
		Password: "otel",
		Endpoint: "localhost:3306",
	})
	sc.client = &mysqlMock

	expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testscrape/expected_metrics.json")
	require.NoError(t, err)

	helper.ScraperTest(t, sc.scrape, expectedFileBytes)
}
