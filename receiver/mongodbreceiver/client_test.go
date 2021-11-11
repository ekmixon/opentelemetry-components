package mongodbreceiver

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/zap"
)

// var _ mongoClient = (*fakeClient)(nil)

type fakeClient struct {
	mock.Mock
}

func (c *fakeClient) Disconnect(ctx context.Context) error {
	args := c.Called(ctx)
	return args.Error(0)
}

func (c *fakeClient) Query(ctx context.Context, database string, command bson.M) (bson.M, error) {
	args := c.Called(ctx, database, command)
	return args.Get(0).(bson.M), args.Error(1)
}

func (c *fakeClient) ListDatabaseNames(ctx context.Context, filter interface{}, options ...*options.ListDatabasesOptions) ([]string, error) {
	args := c.Called(ctx, filter, options)
	return args.Get(0).([]string), args.Error(1)
}

func (c *fakeClient) Connect(ctx context.Context) error {
	args := c.Called(ctx)
	return args.Error(0)
}

func (c *fakeClient) Ping(ctx context.Context, rp *readpref.ReadPref) error {
	args := c.Called(ctx, rp)
	return args.Error(0)
}

func (c *fakeClient) Database(name string, opts ...*options.DatabaseOptions) *mongo.Database {
	args := c.Called(name, opts)
	return args.Get(0).(*mongo.Database)
}

func (c *fakeClient) TestBadClientBadEndpoint(t *testing.T) {
	_ = NewClient(&Config{
		ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
			CollectionInterval: 10 * time.Second,
		},
		TCPAddr: confignet.TCPAddr{
			Endpoint: "ht13a://notvalid:9090",
		},
		Username: "not valid",
		Password: "not valid",
		Timeout:  1 * time.Second,
	}, zap.NewNop())

	// require.Error(t, err)
}

func TestBadClientUnreachable(t *testing.T) {
	cl := NewClient(&Config{
		ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
			CollectionInterval: 10 * time.Second,
		},
		TCPAddr: confignet.TCPAddr{
			Endpoint: "http://notvalid:9090",
		},
		Username: "not valid",
		Password: "not valid",
		Timeout:  1 * time.Second,
	}, zap.NewNop())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := cl.Connect(ctx)
	require.NotNil(t, err)
}

func TestInitClientBadHost(t *testing.T) {
	fakeMongoClient := &fakeClient{}

	client := mongodbClient{
		username: "admin",
		password: "password",
		endpoint: "x:localhost:27017:another_uri",
		client:   fakeMongoClient,
		logger:   zap.NewNop(),
	}

	err := client.initClient()
	require.Error(t, err)
}

func TestDisconnectSuccess(t *testing.T) {
	fakeMongoClient := &fakeClient{}
	fakeMongoClient.On("Disconnect", mock.Anything).Return(nil)

	client := mongodbClient{
		client: fakeMongoClient,
		logger: zap.NewNop(),
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.Disconnect(ctx)
	require.NoError(t, err)
}

func TestDisconnectFailure(t *testing.T) {
	fakeMongoClient := &fakeClient{}
	fakeMongoClient.On("Disconnect", mock.Anything).Return(fmt.Errorf("connection terminated by peer"))

	client := mongodbClient{
		client: fakeMongoClient,
		logger: zap.NewNop(),
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.Disconnect(ctx)
	require.Error(t, err)
}
