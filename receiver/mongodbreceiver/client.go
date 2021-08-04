package mongodbreceiver

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.uber.org/zap"
)

type mongodbClient struct {
	*mongo.Client
	logger  *zap.Logger
	timeout time.Duration
}

func (r *mongodbScraper) initClient(ctx context.Context, logger *zap.Logger, timeout time.Duration) (*mongodbClient, error) {
	authentication := ""
	if r.config.Username != "" && r.config.Password != "" {
		authentication = fmt.Sprintf("%s:%s@", r.config.Username, r.config.Password)
	}

	uri := fmt.Sprintf("mongodb://%s%s", authentication, r.config.Endpoint)

	timeoutCtx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	client, err := mongo.Connect(timeoutCtx, options.Client().ApplyURI(uri))
	return &mongodbClient{
		Client:  client,
		logger:  logger,
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
