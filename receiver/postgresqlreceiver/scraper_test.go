package postgresqlreceiver

import (
	"io/ioutil"
	"testing"

	"github.com/observiq/opentelemetry-components/receiver/helper"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestScraper(t *testing.T) {
	sc := newPostgreSQLScraper(zap.NewNop(), &Config{Databases: []string{"otel"}})
	// Mock the initializeClient function
	initializeClient = func(p *postgreSQLScraper, database string) (client, error) {
		return &fakeClient{database: database, databases: []string{"otel"}}, nil
	}

	expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testscraper/otel/expected_metrics.json")
	require.NoError(t, err)

	helper.ScraperTest(t, sc.scrape, expectedFileBytes)
}

func TestScraperNoDatabaseSingle(t *testing.T) {
	sc := newPostgreSQLScraper(zap.NewNop(), &Config{})
	// Mock the initializeClient function
	initializeClient = func(p *postgreSQLScraper, database string) (client, error) {
		return &fakeClient{database: database, databases: []string{"otel"}}, nil
	}

	expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testscraper/otel/expected_metrics.json")
	require.NoError(t, err)

	helper.ScraperTest(t, sc.scrape, expectedFileBytes)
}

func TestScraperNoDatabaseMultiple(t *testing.T) {
	sc := newPostgreSQLScraper(zap.NewNop(), &Config{})
	// Mock the initializeClient function
	initializeClient = func(p *postgreSQLScraper, database string) (client, error) {
		return &fakeClient{database: database, databases: []string{"otel", "open", "telemetry"}}, nil
	}

	expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testscraper/multiple/expected_metrics.json")
	require.NoError(t, err)

	helper.ScraperTest(t, sc.scrape, expectedFileBytes)
}
