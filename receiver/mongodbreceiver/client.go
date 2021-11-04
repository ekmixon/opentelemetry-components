package mongodbreceiver

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"go.uber.org/zap"
)

type Client interface {
	Query(context.Context, string, bson.M) (bson.M, error)
	ListDatabaseNames(context.Context, interface{}, ...*options.ListDatabasesOptions) ([]string, error)
	Disconnect(context.Context) error
	Connect(context.Context) error
	Ping(ctx context.Context, rp *readpref.ReadPref) error
}

var _ Client = (*mongodbClient)(nil)

type mongodbClient struct {
	*mongo.Client
	logger  *zap.Logger
	timeout time.Duration
}

type buildClient func(config *Config, logger *zap.Logger) (Client, error)

func createClient(config *Config, logger *zap.Logger) (Client, error) {
	authentication := ""
	if config.Username != "" && config.Password != "" {
		authentication = fmt.Sprintf("%s:%s@", config.Username, config.Password)
	}

	uri := fmt.Sprintf("mongodb://%s%s", authentication, config.Endpoint)

	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	return &mongodbClient{
		Client:  client,
		logger:  logger,
		timeout: config.Timeout,
	}, err
}

func (c *mongodbClient) Query(ctx context.Context, database string, command bson.M) (bson.M, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	result := c.Database(database).RunCommand(timeoutCtx, command)

	var document bson.M
	err := result.Decode(&document)
	return document, err
}
