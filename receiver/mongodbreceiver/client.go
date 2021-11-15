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

// Client is an interface that exposes functionality towards a mongo instance
type Client interface {
	ListDatabaseNames(context.Context, interface{}, ...*options.ListDatabasesOptions) ([]string, error)
	Disconnect(context.Context) error
	Connect(context.Context) error
	Ping(ctx context.Context, rp *readpref.ReadPref) error
	Query(context.Context, string, bson.M) (bson.M, error)
}

// mongoClient is the underlying mongo.Client that Client invokes to actually do actions
// against the mongo instance
type mongoClient interface {
	Database(string, ...*options.DatabaseOptions) *mongo.Database
	ListDatabaseNames(context.Context, interface{}, ...*options.ListDatabasesOptions) ([]string, error)
	Connect(context.Context) error
	Disconnect(context.Context) error
	Ping(ctx context.Context, rp *readpref.ReadPref) error
}

type mongodbClient struct {
	// underlying mongo driver client
	client   mongoClient
	username string
	password string
	endpoint string
	logger   *zap.Logger
	timeout  time.Duration
}

// NewClient creates a new client to connect and query mongo for the
// monodbreceiver
func NewClient(config *Config, logger *zap.Logger) Client {
	return &mongodbClient{
		endpoint: config.Endpoint,
		username: config.Username,
		password: config.Password,
		timeout:  config.Timeout,
		logger:   logger,
	}
}

// Connect establishes a connection to mongodb instance
func (c *mongodbClient) Connect(ctx context.Context) error {
	if err := c.ensureClient(ctx); err != nil {
		return fmt.Errorf("unable to instantiate mongo client: %w", err)
	}
	return nil
}

// Disconnect closes attempts to close any open connections
func (c *mongodbClient) Disconnect(ctx context.Context) error {
	if c.client != nil {
		c.logger.Info("disconnecting from mongo server at: " + c.endpoint)
		return c.client.Disconnect(ctx)
	}
	return nil
}

// ListDatabaseNames gets a list of the database name given a filter
func (c *mongodbClient) ListDatabaseNames(ctx context.Context, filters interface{}, options ...*options.ListDatabasesOptions) ([]string, error) {
	return c.client.ListDatabaseNames(ctx, filters, options...)
}

// Ping validates that the connection is truly reachable. Relies on connection to be established via `Connect()`
func (c *mongodbClient) Ping(ctx context.Context, rp *readpref.ReadPref) error {
	return c.client.Ping(ctx, rp)
}

// Query executes a query against a database. Relies on connection to be established via `Connect()`
func (c *mongodbClient) Query(ctx context.Context, database string, command bson.M) (bson.M, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	result := c.client.Database(database).RunCommand(timeoutCtx, command)

	var document bson.M
	err := result.Decode(&document)
	return document, err
}

func (c *mongodbClient) ensureClient(ctx context.Context) error {
	if c.client == nil {
		client, err := mongo.Connect(ctx,
			options.Client().
				ApplyURI(uri(c.username, c.password, c.endpoint)))
		if err != nil {
			return fmt.Errorf("error creating mongo client: %w", err)
		}

		if err := client.Ping(ctx, nil); err != nil {
			return fmt.Errorf("could not connect to %s: %w", c.endpoint, err)
		}

		c.logger.Info(fmt.Sprintf("Mongo connection established to: %s", c.endpoint))
		c.client = client
	}
	return nil
}

func uri(username, password, endpoint string) string {
	return fmt.Sprintf("mongodb://%s%s", authenticationString(username, password), endpoint)
}

func authenticationString(username, password string) string {
	if username != "" && password != "" {
		return fmt.Sprintf("%s:%s@", username, password)
	}
	return ""
}
