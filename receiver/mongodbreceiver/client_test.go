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
	"go.opentelemetry.io/collector/config/configtls"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/zap"
)

var _ mongoClient = (*fakeClient)(nil)

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

func TestValidClient(t *testing.T) {
	client, err := NewClient(&Config{
		ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
			CollectionInterval: 10 * time.Second,
		},
		TCPAddr: confignet.TCPAddr{
			Endpoint: "localhost:27017",
		},
		Username: "username",
		Password: "password",
		Timeout:  1 * time.Second,
	}, zap.NewNop())

	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestBadTLSClient(t *testing.T) {
	client, err := NewClient(&Config{
		ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
			CollectionInterval: 10 * time.Second,
		},
		TCPAddr: confignet.TCPAddr{
			Endpoint: "localhost:27017",
		},
		TLSClientSetting: configtls.TLSClientSetting{
			TLSSetting: configtls.TLSSetting{
				CAFile:   "/dev/temporal.txt",
				CertFile: "/dev/test.txt",
			},
		},
		Username: "username",
		Password: "password",
		Timeout:  1 * time.Second,
	}, zap.NewNop())

	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid tls configuration")
	require.Nil(t, client)
}

func TestBadClientNonTCPAddr(t *testing.T) {
	cl, err := NewClient(&Config{
		ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
			CollectionInterval: 10 * time.Second,
		},
		TCPAddr: confignet.TCPAddr{
			Endpoint: "/dev/null",
		},
		Timeout: 1 * time.Second,
	}, zap.NewNop())

	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = cl.Connect(ctx)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "unable to instantiate mongo client")
}

func TestInitClientBadEndpoint(t *testing.T) {
	client := mongodbClient{
		username: "admin",
		password: "password",
		endpoint: "x:localhost:27017:another_uri",
		logger:   zap.NewNop(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := client.ensureClient(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "error creating")
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

func TestDisconnectNoClient(t *testing.T) {
	client := mongodbClient{
		logger: zap.NewNop(),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
	defer cancel()

	err := client.Disconnect(ctx)
	require.Error(t, err)
}

func TestSuccessfulPing(t *testing.T) {
	fakeMongoClient := &fakeClient{}
	fakeMongoClient.On("Ping", mock.Anything, mock.Anything).Return(nil)
	client := mongodbClient{
		client: fakeMongoClient,
		logger: zap.NewNop(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := client.Ping(ctx, readpref.PrimaryPreferred())
	require.NoError(t, err)
}
