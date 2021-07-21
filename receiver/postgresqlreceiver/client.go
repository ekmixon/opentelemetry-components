package postgresqlreceiver

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

type client interface {
	Close() error
	getCommits() (*MetricStat, error)
	getRollbacks() (*MetricStat, error)
	getBackends() (*MetricStat, error)
	getDatabaseSize() (*MetricStat, error)
	getDatabaseRows() (*MetricStat, error)
	getDatabaseRowsByTable() ([]*MetricStat, error)
	getBlocksRead() (*MetricStat, error)
	getBlocksReadByTable() ([]*MetricStat, error)
	getOperations() (*MetricStat, error)
	getOperationsByTable() ([]*MetricStat, error)
}

type postgreSQLClient struct {
	client   *sql.DB
	database string
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
		client:   db,
		database: conf.databaseName,
	}, nil
}

func (c *postgreSQLClient) Close() error {
	return c.client.Close()
}

type MetricStat struct {
	database string
	table    string
	stats    map[string]string
}

const GlobalTable = "_global"

func (p *postgreSQLClient) getCommits() (*MetricStat, error) {
	query := fmt.Sprintf("SELECT xact_commit FROM pg_stat_database WHERE datname = '%s';", p.database)
	rows, err := p.client.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metricStat MetricStat
	for rows.Next() {
		var commit string
		if err := rows.Scan(&commit); err != nil {
			return nil, err
		}
		metricStat = MetricStat{
			database: p.database,
			table:    GlobalTable,
			stats:    map[string]string{"xact_commit": commit},
		}
	}
	return &metricStat, nil
}

func (p *postgreSQLClient) getRollbacks() (*MetricStat, error) {
	query := fmt.Sprintf("SELECT xact_rollback FROM pg_stat_database WHERE datname = '%s';", p.database)
	rows, err := p.client.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metricStat MetricStat
	for rows.Next() {
		var rollback string
		if err := rows.Scan(&rollback); err != nil {
			return nil, err
		}
		metricStat = MetricStat{
			database: p.database,
			table:    GlobalTable,
			stats:    map[string]string{"xact_rollback": rollback},
		}

	}
	return &metricStat, nil
}

func (p *postgreSQLClient) getBackends() (*MetricStat, error) {
	query := fmt.Sprintf("SELECT count(*) as count from pg_stat_activity WHERE datname = '%s'", p.database)
	rows, err := p.client.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metricStat MetricStat
	for rows.Next() {
		var count string
		if err := rows.Scan(&count); err != nil {
			return nil, err
		}
		metricStat = MetricStat{
			database: p.database,
			table:    GlobalTable,
			stats:    map[string]string{"count": count},
		}

	}
	return &metricStat, nil
}

func (p *postgreSQLClient) getDatabaseSize() (*MetricStat, error) {
	query := fmt.Sprintf("SELECT pg_database_size('%s') as size;", p.database)
	rows, err := p.client.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metricStat MetricStat
	for rows.Next() {
		var size string
		if err := rows.Scan(&size); err != nil {
			return nil, err
		}

		metricStat = MetricStat{
			database: p.database,
			table:    GlobalTable,
			stats:    map[string]string{"db_size": size},
		}

	}
	return &metricStat, nil
}

func (p *postgreSQLClient) getDatabaseRows() (*MetricStat, error) {
	query := `SELECT coalesce(sum(n_live_tup), 0) AS live, 
	coalesce(sum(n_dead_tup), 0) AS dead 
	FROM pg_stat_user_tables;`
	rows, err := p.client.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metricStat MetricStat
	for rows.Next() {
		var live, dead string
		if err := rows.Scan(&live, &dead); err != nil {
			return nil, err
		}
		metricStat = MetricStat{
			database: p.database,
			table:    GlobalTable,
			stats:    map[string]string{"live": live, "dead": dead},
		}
	}
	return &metricStat, nil
}

func (p *postgreSQLClient) getDatabaseRowsByTable() ([]*MetricStat, error) {
	query := `SELECT schemaname, relname,
	n_live_tup AS live, n_dead_tup AS dead 
	FROM pg_stat_user_tables;`
	rows, err := p.client.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metricStats := []*MetricStat{}
	for rows.Next() {
		var schemaname, relname, live, dead string
		if err := rows.Scan(&schemaname, &relname, &live, &dead); err != nil {
			return nil, err
		}
		metricStat := MetricStat{
			database: p.database,
			table:    fmt.Sprintf("%s.%s", schemaname, relname),
			stats:    map[string]string{"live": live, "dead": dead},
		}
		metricStats = append(metricStats, &metricStat)
	}
	return metricStats, nil
}

func (p *postgreSQLClient) getBlocksRead() (*MetricStat, error) {
	query := `SELECT coalesce(sum(heap_blks_read), 0) AS heap_read, 
	coalesce(sum(heap_blks_hit), 0) AS heap_hit, 
	coalesce(sum(idx_blks_read), 0) AS idx_read, 
	coalesce(sum(idx_blks_hit), 0) AS idx_hit, 
	coalesce(sum(toast_blks_read), 0) AS toast_read, 
	coalesce(sum(toast_blks_hit), 0) AS toast_hit, 
	coalesce(sum(tidx_blks_read), 0) AS tidx_read, 
	coalesce(sum(tidx_blks_hit), 0) AS tidx_hit 
	FROM pg_statio_user_tables;`
	rows, err := p.client.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metricStat MetricStat
	for rows.Next() {
		stats := map[string]string{}
		var heapRead, heapHit, idxRead, idxHit, toastRead, toastHit, tidxRead, tidxHit string
		if err := rows.Scan(&heapRead, &heapHit, &idxRead, &idxHit, &toastRead, &toastHit, &tidxRead, &tidxHit); err != nil {
			return nil, err
		}
		stats["heap_read"] = heapRead
		stats["heap_hit"] = heapHit
		stats["idx_read"] = idxRead
		stats["idx_hit"] = idxHit
		stats["toast_read"] = toastRead
		stats["toast_hit"] = toastHit
		stats["tidx_read"] = tidxRead
		stats["tidx_hit"] = tidxHit
		metricStat = MetricStat{
			database: p.database,
			table:    GlobalTable,
			stats:    stats,
		}
	}
	return &metricStat, nil
}

func (p *postgreSQLClient) getBlocksReadByTable() ([]*MetricStat, error) {
	query := `SELECT schemaname, relname, 
	coalesce(heap_blks_read, 0) AS heap_read, 
	coalesce(heap_blks_hit, 0) AS heap_hit, 
	coalesce(idx_blks_read, 0) AS idx_read, 
	coalesce(idx_blks_hit, 0) AS idx_hit, 
	coalesce(toast_blks_read, 0) AS toast_read, 
	coalesce(toast_blks_hit, 0) AS toast_hit, 
	coalesce(tidx_blks_read, 0) AS tidx_read, 
	coalesce(tidx_blks_hit, 0) AS tidx_hit 
	FROM pg_statio_user_tables;`
	rows, err := p.client.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metricStats := []*MetricStat{}
	for rows.Next() {
		stats := map[string]string{}
		var schemaname, relname, heapRead, heapHit, idxRead, idxHit, toastRead, toastHit, tidxRead, tidxHit string
		if err := rows.Scan(&schemaname, &relname, &heapRead, &heapHit, &idxRead, &idxHit, &toastRead, &toastHit, &tidxRead, &tidxHit); err != nil {
			return nil, err
		}
		stats["heap_read"] = heapRead
		stats["heap_hit"] = heapHit
		stats["idx_read"] = idxRead
		stats["idx_hit"] = idxHit
		stats["toast_read"] = toastRead
		stats["toast_hit"] = toastHit
		stats["tidx_read"] = tidxRead
		stats["tidx_hit"] = tidxHit

		metricStat := MetricStat{
			database: p.database,
			table:    fmt.Sprintf("%s.%s", schemaname, relname),
			stats:    stats,
		}
		metricStats = append(metricStats, &metricStat)
	}
	return metricStats, nil
}

func (p *postgreSQLClient) getOperations() (*MetricStat, error) {
	query := `SELECT coalesce(sum(seq_scan), 0) AS seq, 
	coalesce(sum(seq_tup_read), 0) AS seq_tup_read, 
	coalesce(sum(idx_scan), 0) AS idx, 
	coalesce(sum(idx_tup_fetch), 0) AS idx_tup_fetch 
	FROM pg_stat_user_tables;`
	rows, err := p.client.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metricStat MetricStat
	for rows.Next() {
		stats := map[string]string{}
		var seq, seq_tup_read, idx, idx_tup_fetch string
		if err := rows.Scan(&seq, &seq_tup_read, &idx, &idx_tup_fetch); err != nil {
			return nil, err
		}
		stats["seq"] = seq
		stats["seq_tup_read"] = seq_tup_read
		stats["idx"] = idx
		stats["idx_tup_fetch"] = idx_tup_fetch

		metricStat = MetricStat{
			database: p.database,
			table:    GlobalTable,
			stats:    stats,
		}

	}
	return &metricStat, nil
}

func (p *postgreSQLClient) getOperationsByTable() ([]*MetricStat, error) {
	query := `SELECT schemaname, relname,
	coalesce(seq_scan, 0) AS seq,
	coalesce(seq_tup_read, 0) AS seq_tup_read,
	coalesce(idx_scan, 0) AS idx,
	coalesce(idx_tup_fetch, 0) AS idx_tup_fetch
	FROM pg_stat_user_tables;`
	rows, err := p.client.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metricStats := []*MetricStat{}
	for rows.Next() {
		stats := map[string]string{}
		var schemaname, relname, seq, seq_tup_read, idx, idx_tup_fetch string
		if err := rows.Scan(&schemaname, &relname, &seq, &seq_tup_read, &idx, &idx_tup_fetch); err != nil {
			return nil, err
		}
		stats["seq"] = seq
		stats["seq_tup_read"] = seq_tup_read
		stats["idx"] = idx
		stats["idx_tup_fetch"] = idx_tup_fetch

		metricStat := MetricStat{
			database: p.database,
			table:    fmt.Sprintf("%s.%s", schemaname, relname),
			stats:    stats,
		}
		metricStats = append(metricStats, &metricStat)
	}
	return metricStats, nil
}
