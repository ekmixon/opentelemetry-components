package mongodbreceiver

import (
	"io/ioutil"
	"net"
	"testing"

	"github.com/observiq/opentelemetry-components/receiver/helper"
	"github.com/observiq/opentelemetry-components/receiver/mongodbreceiver/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

func TestScrape(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = net.JoinHostPort("localhost", "37017")
	cfg.Username = "otel"
	cfg.Password = "otel"
	databaseNames := []string{"admin", "config", "local"}

	mockClient := &mocks.Client{}
	mockClient.On("Connect", mock.Anything).Return(nil)
	mockClient.On("Ping", mock.Anything, readpref.PrimaryPreferred()).Return(nil)
	mockClient.On("Disconnect", mock.Anything).Return(nil)
	mockClient.On("ListDatabaseNames", mock.Anything, bson.D{}).Return(databaseNames, nil)

	createClient := func(config *Config, logger *zap.Logger) (Client, error) {
		return mockClient, nil
	}

	scraper := mongodbScraper{
		logger:      zap.NewNop(),
		config:      cfg,
		buildClient: createClient,
	}

	expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testscraper/expected_metrics.json")
	require.NoError(t, err)

	helper.ScraperTest(t, scraper.scrape, expectedFileBytes)
}
