package mongodbreceiver

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"testing"

	"github.com/observiq/opentelemetry-components/receiver/helper"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.opentelemetry.io/collector/config/configtls"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = scraper.shutdown(ctx)
	require.NoError(t, err)
}

func TestCreateNewScraper(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = net.JoinHostPort("localhost", "37017")
	cfg.Username = "otel"
	cfg.Password = "otel"
	cfg.TLSClientSetting = configtls.TLSClientSetting{
		InsecureSkipVerify: true,
	}

	scraper := newMongodbScraper(zap.NewNop(), cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := scraper.start(ctx, nil)
	require.NoError(t, err)

	require.True(t, scraper.config.TLSClientSetting.InsecureSkipVerify)
}

func TestScraperNoClient(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = net.JoinHostPort("localhost", "37017")
	cfg.Username = "otel"
	cfg.Password = "otel"

	scraper := newMongodbScraper(zap.NewNop(), cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := scraper.scrape(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no client was initialized before calling scrape")

	err = scraper.shutdown(ctx)
	require.NoError(t, err)
}

func TestScrapeClientConnectionFailure(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = net.JoinHostPort("localhost", "37017")
	cfg.Username = "otel"
	cfg.Password = "otel"

	client := &fakeClient{}
	client.On("Connect", mock.Anything).Return(fmt.Errorf("Could not connect"))
	client.On("Disconnect", mock.Anything).Return(nil)

	scraper := &mongodbScraper{
		config: cfg,
		logger: zap.NewNop(),
		client: client,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resources, err := scraper.scrape(ctx)
	require.Error(t, err)
	require.Equal(t, resources.Len(), 0)
}

func TestScrapeClientPingFailure(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = net.JoinHostPort("localhost", "37017")
	cfg.Username = "otel"
	cfg.Password = "otel"

	client := &fakeClient{}
	client.On("Connect", mock.Anything).Return(nil)
	client.On("Ping", mock.Anything, mock.Anything).Return(fmt.Errorf("could not ping"))
	client.On("Disconnect", mock.Anything).Return(nil)

	scraper := &mongodbScraper{
		config: cfg,
		logger: zap.NewNop(),
		client: client,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resources, err := scraper.scrape(ctx)
	require.Error(t, err)
	require.Equal(t, resources.Len(), 0)
}

func BenchmarkScrape(b *testing.B) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = net.JoinHostPort("localhost", "37017")
	cfg.Username = "otel"
	cfg.Password = "otel"

	databaseList := []string{"test-database"}

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

	for n := 0; n < b.N; n++ {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		_, _ = scraper.scrape(ctx)
	}
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
