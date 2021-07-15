package postgresqlreceiver

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

type client interface {
	Closed() bool
	Close() error
}

type postgreSQLClient struct {
	client *sql.DB
	closed bool
}

var _ client = (*postgreSQLClient)(nil)

type postgreSQLConfig struct {
	username     string
	password     string
	databaseName string
	endpoint     string
}

func newPostgreSQLClient(conf postgreSQLConfig) (*postgreSQLClient, error) {
	endpoint := strings.Split(conf.endpoint, ":")
	connStr := fmt.Sprintf("port=%s host=%s user=%s password=%s dbname=%s sslmode=disable", endpoint[1], endpoint[0], conf.username, conf.password, conf.databaseName)

	conn, err := pq.NewConnector(connStr)
	if err != nil {
		return nil, err
	}

	db := sql.OpenDB(conn)

	return &postgreSQLClient{
		client: db,
	}, nil
}

func (c *postgreSQLClient) Closed() bool {
	return c.closed
}

func (c *postgreSQLClient) Close() error {
	err := c.client.Close()
	if err != nil {
		return err
	}
	c.closed = true
	return nil
}
