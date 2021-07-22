package postgresqlreceiver

var _ client = (*fakeClient)(nil)

type fakeClient struct {
	database string
}

func (c *fakeClient) Close() error {
	return nil
}

func (c *fakeClient) getCommits() (*MetricStat, error) {
	return &MetricStat{
		database: c.database,
		table:    GlobalTable,
		stats:    map[string]string{"xact_commit": "1"},
	}, nil
}

func (c *fakeClient) getRollbacks() (*MetricStat, error) {
	return &MetricStat{
		database: c.database,
		table:    GlobalTable,
		stats:    map[string]string{"xact_rollback": "2"},
	}, nil
}

func (c *fakeClient) getBackends() (*MetricStat, error) {
	return &MetricStat{
		database: c.database,
		table:    GlobalTable,
		stats:    map[string]string{"count": "3"},
	}, nil
}

func (c *fakeClient) getDatabaseSize() (*MetricStat, error) {
	return &MetricStat{
		database: c.database,
		table:    GlobalTable,
		stats:    map[string]string{"db_size": "4"},
	}, nil
}

func (c *fakeClient) getDatabaseRows() (*MetricStat, error) {
	return &MetricStat{
		database: c.database,
		table:    GlobalTable,
		stats:    map[string]string{"live": "5", "dead": "6"},
	}, nil
}

func (c *fakeClient) getDatabaseRowsByTable() ([]*MetricStat, error) {
	metricStats := []*MetricStat{}
	metricStats = append(metricStats, &MetricStat{
		database: c.database,
		table:    "table1",
		stats:    map[string]string{"live": "7", "dead": "8"},
	})
	metricStats = append(metricStats, &MetricStat{
		database: c.database,
		table:    "table2",
		stats:    map[string]string{"live": "9", "dead": "10"},
	})
	return metricStats, nil
}

func (c *fakeClient) getBlocksRead() (*MetricStat, error) {
	return &MetricStat{
		database: c.database,
		table:    GlobalTable,
		stats: map[string]string{
			"heap_read":  "11",
			"heap_hit":   "12",
			"idx_read":   "13",
			"idx_hit":    "14",
			"toast_read": "15",
			"toast_hit":  "16",
			"tidx_read":  "17",
			"tidx_hit":   "18",
		},
	}, nil
}

func (c *fakeClient) getBlocksReadByTable() ([]*MetricStat, error) {
	metricStats := []*MetricStat{}
	metricStats = append(metricStats, &MetricStat{
		database: c.database,
		table:    "table1",
		stats: map[string]string{
			"heap_read":  "19",
			"heap_hit":   "20",
			"idx_read":   "21",
			"idx_hit":    "22",
			"toast_read": "23",
			"toast_hit":  "24",
			"tidx_read":  "25",
			"tidx_hit":   "26",
		},
	})
	metricStats = append(metricStats, &MetricStat{
		database: c.database,
		table:    "table2",
		stats: map[string]string{
			"heap_read":  "27",
			"heap_hit":   "28",
			"idx_read":   "29",
			"idx_hit":    "30",
			"toast_read": "31",
			"toast_hit":  "32",
			"tidx_read":  "33",
			"tidx_hit":   "34",
		},
	})
	return metricStats, nil
}

func (c *fakeClient) getOperations() (*MetricStat, error) {
	return &MetricStat{
		database: c.database,
		table:    GlobalTable,
		stats: map[string]string{
			"seq":           "35",
			"seq_tup_read":  "36",
			"idx":           "37",
			"idx_tup_fetch": "38",
		},
	}, nil
}

func (c *fakeClient) getOperationsByTable() ([]*MetricStat, error) {
	metricStats := []*MetricStat{}
	metricStats = append(metricStats, &MetricStat{
		database: c.database,
		table:    "table1",
		stats: map[string]string{
			"seq":           "39",
			"seq_tup_read":  "40",
			"idx":           "41",
			"idx_tup_fetch": "42",
		},
	})
	metricStats = append(metricStats, &MetricStat{
		database: c.database,
		table:    "table2",
		stats: map[string]string{
			"seq":           "43",
			"seq_tup_read":  "44",
			"idx":           "45",
			"idx_tup_fetch": "46",
		},
	})
	return metricStats, nil
}
