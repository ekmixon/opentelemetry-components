package mongodbreceiver

import (
	"context"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/observiq/opentelemetry-components/receiver/helper"
	"github.com/observiq/opentelemetry-components/receiver/mongodbreceiver/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.opentelemetry.io/collector/component/componenttest"
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

var (
	databaseNames     = []string{"admin", "config", "local"}
	serverStatusQuery = bson.M{"serverStatus": 1}
	dbStatusQuery     = bson.M{"dbStats": 1}
)

func TestScrape(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)

	mockClient := &mocks.Client{}
	mockClient.On("Connect", mock.Anything).Return(nil)
	mockClient.On("Ping", mock.Anything, readpref.PrimaryPreferred()).Return(nil)
	mockClient.On("Disconnect", mock.Anything).Return(nil)
	mockClient.On("ListDatabaseNames", mock.Anything, bson.D{}).Return(databaseNames, nil)

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

	expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testscraper/multiple_databases_metrics.json")
	require.NoError(t, err)
	helper.ScraperTest(t, scraper.scrape, expectedFileBytes)
}

func TestScrapeErrorBuildClient(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)

	obs, _ := observer.New(zap.ErrorLevel)

	createClient := func(config *Config, logger *zap.Logger) (Client, error) {
		return nil, errors.New("buildClient error")
	}

	scraper := mongodbScraper{
		logger:      zap.New(obs),
		config:      cfg,
		buildClient: createClient,
	}

	rms, err := scraper.scrape(context.Background())
	require.EqualError(t, errors.New("buildClient error"), err.Error())
	require.Equal(t, pdata.NewResourceMetricsSlice(), rms)
}

func TestScrapeErrorConnect(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)

	mockClient := &mocks.Client{}
	mockClient.On("Connect", mock.Anything).Return(errors.New("connection error"))

	obs, _ := observer.New(zap.ErrorLevel)

	createClient := func(config *Config, logger *zap.Logger) (Client, error) {
		return mockClient, nil
	}

	scraper := mongodbScraper{
		logger:      zap.New(obs),
		config:      cfg,
		buildClient: createClient,
	}

	rms, err := scraper.scrape(context.Background())
	require.EqualError(t, errors.New("connection error"), err.Error())
	require.Equal(t, pdata.NewResourceMetricsSlice(), rms)
}

func TestScrapeErrorPing(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)

	mockClient := &mocks.Client{}
	mockClient.On("Connect", mock.Anything).Return(nil)
	mockClient.On("Ping", mock.Anything, readpref.PrimaryPreferred()).Return(errors.New("ping error"))
	mockClient.On("Disconnect", mock.Anything).Return(nil)

	obs, _ := observer.New(zap.ErrorLevel)

	createClient := func(config *Config, logger *zap.Logger) (Client, error) {
		return mockClient, nil
	}

	scraper := mongodbScraper{
		logger:      zap.New(obs),
		config:      cfg,
		buildClient: createClient,
	}

	rms, err := scraper.scrape(context.Background())
	require.EqualError(t, errors.New("ping error"), err.Error())
	require.Equal(t, pdata.NewResourceMetricsSlice(), rms)
}

func TestScrapeErrorListDatabaseNames(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)

	mockClient := &mocks.Client{}
	mockClient.On("Connect", mock.Anything).Return(nil)
	mockClient.On("Ping", mock.Anything, readpref.PrimaryPreferred()).Return(nil)
	mockClient.On("Disconnect", mock.Anything).Return(nil)
	mockClient.On("ListDatabaseNames", mock.Anything, bson.D{}).Return(nil, errors.New("list database names error"))

	obs, _ := observer.New(zap.ErrorLevel)

	createClient := func(config *Config, logger *zap.Logger) (Client, error) {
		return mockClient, nil
	}

	scraper := mongodbScraper{
		logger:      zap.New(obs),
		config:      cfg,
		buildClient: createClient,
	}

	rms, err := scraper.scrape(context.Background())
	require.EqualError(t, errors.New("list database names error"), err.Error())
	require.Equal(t, pdata.NewResourceMetricsSlice(), rms)
}

func TestScrapeErrorDisconnect(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)

	mockClient := &mocks.Client{}
	mockClient.On("Connect", mock.Anything).Return(nil)
	mockClient.On("Ping", mock.Anything, readpref.PrimaryPreferred()).Return(nil)
	mockClient.On("ListDatabaseNames", mock.Anything, bson.D{}).Return(databaseNames, nil)

	mockClient.On("Query", mock.Anything, databaseNames[0], serverStatusQuery).Return(nil, nil)

	// admin
	mockClient.On("Query", mock.Anything, databaseNames[0], dbStatusQuery).Return(bson.M{}, nil)
	mockClient.On("Query", mock.Anything, databaseNames[0], serverStatusQuery).Return(bson.M{}, nil)

	// config
	mockClient.On("Query", mock.Anything, databaseNames[1], dbStatusQuery).Return(bson.M{}, nil)
	mockClient.On("Query", mock.Anything, databaseNames[1], serverStatusQuery).Return(bson.M{}, nil)

	// local
	mockClient.On("Query", mock.Anything, databaseNames[2], dbStatusQuery).Return(bson.M{}, nil)
	mockClient.On("Query", mock.Anything, databaseNames[2], serverStatusQuery).Return(bson.M{}, nil)

	mockClient.On("Disconnect", mock.Anything).Return(errors.New("disconnect error"))

	obs, _ := observer.New(zap.ErrorLevel)
	// TODO: test logger when defer Disconnect

	createClient := func(config *Config, logger *zap.Logger) (Client, error) {
		return mockClient, nil
	}

	scraper := mongodbScraper{
		logger:      zap.New(obs),
		config:      cfg,
		buildClient: createClient,
	}

	expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testscraper/empty_metrics.json")
	require.NoError(t, err)

	helper.ScraperTest(t, scraper.scrape, expectedFileBytes)
}

func TestScrapeNoErrEmptyMetricsErrorLogs(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)

	mockClient := &mocks.Client{}
	mockClient.On("Connect", mock.Anything).Return(nil)
	mockClient.On("Ping", mock.Anything, readpref.PrimaryPreferred()).Return(nil)
	mockClient.On("Disconnect", mock.Anything).Return(nil)
	mockClient.On("ListDatabaseNames", mock.Anything, bson.D{}).Return(databaseNames, nil)

	mockClient.On("Query", mock.Anything, databaseNames[0], serverStatusQuery).Return(nil, errors.New("admin specialMetrics error"))

	// admin
	mockClient.On("Query", mock.Anything, databaseNames[0], dbStatusQuery).Return(bson.M{}, errors.New("admin dbstats error"))
	mockClient.On("Query", mock.Anything, databaseNames[0], serverStatusQuery).Return(bson.M{}, errors.New("admin serverStatus error"))

	// config
	mockClient.On("Query", mock.Anything, databaseNames[1], dbStatusQuery).Return(bson.M{}, errors.New("config dbstats error"))
	mockClient.On("Query", mock.Anything, databaseNames[1], serverStatusQuery).Return(bson.M{}, errors.New("config serverStatus error"))

	// local
	mockClient.On("Query", mock.Anything, databaseNames[2], dbStatusQuery).Return(bson.M{}, errors.New("local dbstats error"))
	mockClient.On("Query", mock.Anything, databaseNames[2], serverStatusQuery).Return(bson.M{}, errors.New("local serverStatus error"))

	createClient := func(config *Config, logger *zap.Logger) (Client, error) {
		return mockClient, nil
	}

	obs, _ := observer.New(zap.ErrorLevel)

	scraper := mongodbScraper{
		logger:      zap.New(obs),
		config:      cfg,
		buildClient: createClient,
	}

	expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testscraper/empty_metrics.json")
	require.NoError(t, err)

	helper.ScraperTest(t, scraper.scrape, expectedFileBytes)
}

func TestStart(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	scraper := mongodbScraper{
		logger:      zap.NewNop(),
		config:      cfg,
		buildClient: nil,
	}

	err := scraper.start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)
}

func TestGetIntMetric(t *testing.T) {
	testCases := []struct {
		desc           string
		doc            bson.M
		path           []string
		expectedResult int64
		expectedErr    error
	}{
		{
			desc:           "no error parse int",
			doc:            bson.M{"$numberInt": 1},
			path:           []string{"$numberInt"},
			expectedResult: 1,
			expectedErr:    nil,
		},
		{
			desc:           "no error parse string",
			doc:            bson.M{"$numberInt": "1"},
			path:           []string{"$numberInt"},
			expectedResult: 1,
			expectedErr:    nil,
		},
		{
			desc:           "no error parse int32",
			doc:            bson.M{"$number": int32(1)},
			path:           []string{"$number"},
			expectedResult: 1,
			expectedErr:    nil,
		},
		{
			desc:           "no error parse int64",
			doc:            bson.M{"$number": int64(1)},
			path:           []string{"$number"},
			expectedResult: 1,
			expectedErr:    nil,
		},
		{
			desc:           "no error recursive call",
			doc:            bson.M{"key": bson.M{"$number": 1}},
			path:           []string{"key", "$number"},
			expectedResult: 1,
			expectedErr:    nil,
		},
		{
			desc:        "error parse double",
			doc:         bson.M{"$number": 1.0},
			path:        []string{"$number"},
			expectedErr: errors.New("unexpected type found when parsing int:"),
		},
		{
			desc:        "error parse string",
			doc:         bson.M{"$numberInt": "one"},
			path:        []string{"$numberInt"},
			expectedErr: errors.New("invalid syntax"),
		},
		{
			desc:        "error parse nil",
			doc:         bson.M{"$numberInt": nil},
			path:        []string{"$numberInt"},
			expectedErr: errors.New("nil found when digging for metric"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result, err := getIntMetricValue(tc.doc, tc.path)
			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestGetDoubleMetric(t *testing.T) {
	testCases := []struct {
		desc           string
		doc            bson.M
		path           []string
		expectedResult float64
		expectedErr    error
	}{
		{
			desc:           "no error parse float",
			doc:            bson.M{"$numberInt": 1.1},
			path:           []string{"$numberInt"},
			expectedResult: 1.1,
			expectedErr:    nil,
		},
		{
			desc:           "no error parse string",
			doc:            bson.M{"$numberInt": "1.1"},
			path:           []string{"$numberInt"},
			expectedResult: 1.1,
			expectedErr:    nil,
		},
		// {
		// 	desc:           "no error parse float32",
		// 	doc:            bson.M{"$number": float32(1.1)},
		// 	path:           []string{"$number"},
		// 	expectedResult: 1,
		// 	expectedErr:    nil,
		// },
		{
			desc:           "no error parse float64",
			doc:            bson.M{"$number": float64(1.1)},
			path:           []string{"$number"},
			expectedResult: 1.1,
			expectedErr:    nil,
		},
		{
			desc:           "no error recursive call",
			doc:            bson.M{"key": bson.M{"$number": 1.1}},
			path:           []string{"key", "$number"},
			expectedResult: 1.1,
			expectedErr:    nil,
		},
		{
			desc:        "error parse int",
			doc:         bson.M{"$number": 1},
			path:        []string{"$number"},
			expectedErr: errors.New("unexpected type found when parsing double:"),
		},
		{
			desc:        "error parse string",
			doc:         bson.M{"$numberInt": "one"},
			path:        []string{"$numberInt"},
			expectedErr: errors.New("invalid syntax"),
		},
		{
			desc:        "error parse nil",
			doc:         bson.M{"$numberInt": nil},
			path:        []string{"$numberInt"},
			expectedErr: errors.New("nil found when digging for metric"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result, err := getDoubleMetricValue(tc.doc, tc.path)
			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
		})
	}
}
