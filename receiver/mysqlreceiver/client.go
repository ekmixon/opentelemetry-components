package mysqlreceiver

import (
	"database/sql"
	"fmt"

	// registers the mysql driver
	_ "github.com/go-sql-driver/mysql"
)

type client interface {
	getGlobalStats() ([]*Stat, error)
	getInnodbStats() ([]*Stat, error)
	Closed() bool
	Close() error
}

type mySQLClient struct {
	client *sql.DB
	closed bool
}

var _ client = (*mySQLClient)(nil)

type mySQLConfig struct {
	username string
	password string
	database string
	endpoint string
}

func newMySQLClient(conf mySQLConfig) (*mySQLClient, error) {
	connStr := fmt.Sprintf("%s:%s@tcp(%s)/%s", conf.username, conf.password, conf.endpoint, conf.database)

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, err
	}

	return &mySQLClient{
		client: db,
	}, nil
}

// getGlobalStats queries the db for global status metrics.
func (c *mySQLClient) getGlobalStats() ([]*Stat, error) {
	query := "SHOW GLOBAL STATUS;"
	return Query(*c, query)
}

// getInnodbStats queries the db for innodb metrics.
func (c *mySQLClient) getInnodbStats() ([]*Stat, error) {
	query := "SELECT name, count FROM information_schema.innodb_metrics WHERE name LIKE '%buffer_pool_size%';"
	return Query(*c, query)
}

type Stat struct {
	key   string
	value string
}

func Query(c mySQLClient, query string) ([]*Stat, error) {
	rows, err := c.client.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	stats := make([]*Stat, 0)
	for rows.Next() {
		var stat Stat
		if err := rows.Scan(&stat.key, &stat.value); err != nil {
			return nil, err
		}
		stats = append(stats, &stat)
	}

	return stats, nil
}

func (c *mySQLClient) Closed() bool {
	return c.closed
}

func (c *mySQLClient) Close() error {
	err := c.client.Close()
	if err != nil {
		return err
	}
	c.closed = true
	return nil
}
