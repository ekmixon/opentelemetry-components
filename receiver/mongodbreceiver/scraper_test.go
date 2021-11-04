package mongodbreceiver

import (
	"context"
	"errors"
	"io/ioutil"
	"net"
	"testing"

	"github.com/observiq/opentelemetry-components/receiver/helper"
	"github.com/observiq/opentelemetry-components/receiver/mongodbreceiver/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func GetAdmin() bson.M {
	var doc bson.M
	admin, _ := ioutil.ReadFile("./testdata/admin.json")
	_ = bson.UnmarshalExtJSON(admin, true, &doc)

	return doc
}

func GetDbStats() bson.M {
	var doc bson.M
	dbStats, _ := ioutil.ReadFile("./testdata/dbstats.json")
	_ = bson.UnmarshalExtJSON(dbStats, true, &doc)

	return doc
}

func GetServerStatus() bson.M {
	var doc bson.M
	dbStats, _ := ioutil.ReadFile("./testdata/serverstatus.json")
	_ = bson.UnmarshalExtJSON(dbStats, true, &doc)

	return doc
}

func TestScrape(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = net.JoinHostPort("localhost", "37017")
	cfg.Username = "otel"
	cfg.Password = "otel"
	databaseNames := []string{"admin", "config", "local"}
	serverStatusQuery := bson.M{"serverStatus": 1}
	dbStatusQuery := bson.M{"dbStats": 1}

	mockClient := &mocks.Client{}
	mockClient.On("Connect", mock.Anything).Return(nil)
	mockClient.On("Ping", mock.Anything, readpref.PrimaryPreferred()).Return(nil)
	mockClient.On("Disconnect", mock.Anything).Return(nil)
	mockClient.On("ListDatabaseNames", mock.Anything, bson.D{}).Return(databaseNames, nil)

	// mockClient.On("Query", mock.Anything, databaseNames[0], serverStatusQuery).Return(nil, nil)
	mockClient.On("Query", mock.Anything, databaseNames[0], serverStatusQuery).Return(GetAdmin(), nil)

	// admin
	mockClient.On("Query", mock.Anything, databaseNames[0], dbStatusQuery).Return(GetDbStats(), nil)
	mockClient.On("Query", mock.Anything, databaseNames[0], serverStatusQuery).Return(GetServerStatus(), nil)

	// config
	mockClient.On("Query", mock.Anything, databaseNames[1], dbStatusQuery).Return(GetDbStats(), nil)
	mockClient.On("Query", mock.Anything, databaseNames[1], serverStatusQuery).Return(GetServerStatus(), nil)

	// local
	mockClient.On("Query", mock.Anything, databaseNames[2], dbStatusQuery).Return(GetDbStats(), nil)
	mockClient.On("Query", mock.Anything, databaseNames[2], serverStatusQuery).Return(GetServerStatus(), nil)

	createClient := func(config *Config, logger *zap.Logger) (Client, error) {
		return mockClient, nil
	}

	scraper := mongodbScraper{
		logger:      zap.NewNop(),
		config:      cfg,
		buildClient: createClient,
	}

	rms, err := scraper.scrape(context.Background())
	require.NoError(t, err)
	require.NotNil(t, rms)
	// require.EqualValues(t, pdata.ResourceMetricsSlice{}, rms)

	// expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testscraper/expected_metrics.json")
	// require.NoError(t, err)

	// helper.ScraperTest(t, scraper.scrape, expectedFileBytes)
}

func TestScrapeNoErrEmptyMetrics(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = net.JoinHostPort("localhost", "37017")
	cfg.Username = "otel"
	cfg.Password = "otel"
	databaseNames := []string{"admin", "config", "local"}
	serverStatusQuery := bson.M{"serverStatus": 1}
	dbStatusQuery := bson.M{"dbStats": 1}

	mockClient := &mocks.Client{}
	mockClient.On("Connect", mock.Anything).Return(nil)
	mockClient.On("Ping", mock.Anything, readpref.PrimaryPreferred()).Return(nil)
	mockClient.On("Disconnect", mock.Anything).Return(nil)
	mockClient.On("ListDatabaseNames", mock.Anything, bson.D{}).Return(databaseNames, nil)

	// mockClient.On("Query", mock.Anything, databaseNames[0], serverStatusQuery).Return(nil, nil)
	mockClient.On("Query", mock.Anything, databaseNames[0], serverStatusQuery).Return(nil, nil)

	// admin
	mockClient.On("Query", mock.Anything, databaseNames[0], dbStatusQuery).Return(bson.M{}, errors.New("Failed admin"))
	mockClient.On("Query", mock.Anything, databaseNames[0], serverStatusQuery).Return(bson.M{}, nil)

	// config
	mockClient.On("Query", mock.Anything, databaseNames[1], dbStatusQuery).Return(bson.M{}, nil)
	mockClient.On("Query", mock.Anything, databaseNames[1], serverStatusQuery).Return(bson.M{}, nil)

	// local
	mockClient.On("Query", mock.Anything, databaseNames[2], dbStatusQuery).Return(bson.M{}, errors.New("Failed local"))
	mockClient.On("Query", mock.Anything, databaseNames[2], serverStatusQuery).Return(bson.M{}, nil)

	createClient := func(config *Config, logger *zap.Logger) (Client, error) {
		return mockClient, nil
	}

	obs, logs := observer.New(zap.ErrorLevel)

	scraper := mongodbScraper{
		logger:      zap.New(obs),
		config:      cfg,
		buildClient: createClient,
	}

	// require.Equal(t, 1, logs.Len())
	require.Equal(t, []observer.LoggedEntry{
		// {
		// 	Entry: zapcore.Entry{
		// 		Level:   zap.ErrorLevel,
		// 		Message: "test"},
		// 	Context: []zapcore.Field{
		// 		zap.Error(fmt.Errorf("test")),
		// 	},
		// },
	}, logs.AllUntimed())

	// scrapedMetrics, _ := scraper.scrape(context.Background())
	// actualMetrics := pdata.NewMetrics()
	// scrapedMetrics.CopyTo(actualMetrics.ResourceMetrics())

	// bytes, err := otlp.NewJSONMetricsMarshaler().MarshalMetrics(actualMetrics)
	// require.NoError(t, err)
	// err = ioutil.WriteFile("./testdata/examplejsonmetrics/testscraper/empty_metrics.json", bytes, 0644)
	// require.NoError(t, err)

	expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testscraper/empty_metrics.json")
	require.NoError(t, err)

	helper.ScraperTest(t, scraper.scrape, expectedFileBytes)
}

func TestScrapeErrorConnect(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = net.JoinHostPort("localhost", "37017")
	cfg.Username = "otel"
	cfg.Password = "otel"

	mockClient := &mocks.Client{}
	mockClient.On("Connect", mock.Anything).Return(errors.New("connection error"))

	obs, _ := observer.New(zap.ErrorLevel)

	scraper := mongodbScraper{
		logger:      zap.New(obs),
		config:      cfg,
		buildClient: createClient,
	}

	rms, err := scraper.scrape(context.Background())
	require.Error(t, err)
	require.Equal(t, pdata.NewResourceMetricsSlice(), rms)
}

// func TestScrapeErrorDisconnect(t *testing.T) {
// 	f := NewFactory()
// 	cfg := f.CreateDefaultConfig().(*Config)
// 	cfg.Endpoint = net.JoinHostPort("localhost", "37017")
// 	cfg.Username = "otel"
// 	cfg.Password = "otel"

// 	mockClient := &mocks.Client{}
// 	mockClient.On("Disconnect", mock.Anything).Return(errors.New("Disconnect error"))

// 	obs, _ := observer.New(zap.ErrorLevel)

// 	scraper := mongodbScraper{
// 		logger:      zap.New(obs),
// 		config:      cfg,
// 		buildClient: createClient,
// 	}

// 	rms, err := scraper.scrape(context.Background())
// 	require.Error(t, err)
// 	require.Equal(t, pdata.NewResourceMetricsSlice(), rms)
// }
