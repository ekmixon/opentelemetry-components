package mongodbreceiver

import (
	"io/ioutil"
	"net"
	"testing"

	"github.com/observiq/opentelemetry-components/receiver/helper"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

func TestScrape(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = net.JoinHostPort("localhost", "37017")
	cfg.Username = "otel"
	cfg.Password = "otel"

	databaseList := []string{"fakedatabase"}

	client := &fakeClient{}
	client.On("Connect", mock.Anything).Return(nil)
	client.On("Disconnect", mock.Anything).Return(nil)
	client.On("Ping", mock.Anything, mock.Anything).Return(nil)
	client.On("ListDatabaseNames", mock.Anything, mock.Anything, mock.Anything).Return(databaseList, nil)
	client.On("Query", mock.Anything, "admin", mock.Anything).Return(loadAdminData())
	client.On("Query", mock.Anything, mock.Anything, bson.M{"serverStatus": 1}).Return(loadServerStatus())
	client.On("Query", mock.Anything, mock.Anything, bson.M{"dbStats": 1}).Return(loadDBStats())

	scraper := mongodbScraper{
		logger: zap.NewNop(),
		config: cfg,
		client: client,
	}

	expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testscraper/expected_metrics.json")
	require.NoError(t, err)

	helper.ScraperTest(t, scraper.scrape, expectedFileBytes)
}

func loadAdminData() (bson.M, error) {
	var doc bson.M
	admin, err := ioutil.ReadFile("./testdata/admin.json")
	if err != nil {
		return nil, err
	}
	err = bson.UnmarshalExtJSON(admin, true, &doc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func loadDBStats() (bson.M, error) {
	var doc bson.M
	dbStats, err := ioutil.ReadFile("./testdata/dbstats.json")
	if err != nil {
		return nil, err
	}
	err = bson.UnmarshalExtJSON(dbStats, true, &doc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func loadServerStatus() (bson.M, error) {
	var doc bson.M
	serverStatus, err := ioutil.ReadFile("./testdata/serverstatus.json")
	if err != nil {
		return nil, err
	}
	err = bson.UnmarshalExtJSON(serverStatus, true, &doc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}
