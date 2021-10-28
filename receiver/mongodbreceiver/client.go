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

type client interface {
	query(context.Context, string, bson.M) (bson.M, error)
	ListDatabaseNames(context.Context, interface{}, ...*options.ListDatabasesOptions) ([]string, error)
	Disconnect(context.Context) error
	Connect(context.Context) error
	Ping(ctx context.Context, rp *readpref.ReadPref) error
}

var _ client = (*mongodbClient)(nil)

type mongodbClient struct {
	*mongo.Client
	logger  *zap.Logger
	timeout time.Duration
}

func (r *mongodbScraper) initClient(timeout time.Duration) (*mongodbClient, error) {
	authentication := ""
	if r.config.Username != "" && r.config.Password != "" {
		authentication = fmt.Sprintf("%s:%s@", r.config.Username, r.config.Password)
	}

	uri := fmt.Sprintf("mongodb://%s%s", authentication, r.config.Endpoint)

	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	return &mongodbClient{
		Client:  client,
		logger:  r.logger,
		timeout: timeout,
	}, err
}

func (c *mongodbClient) query(ctx context.Context, database string, command bson.M) (bson.M, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	result := c.Database(database).RunCommand(timeoutCtx, command)

	var document bson.M
	err := result.Decode(&document)
	return document, err
}
