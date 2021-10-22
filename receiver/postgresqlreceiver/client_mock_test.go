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
		stats:    map[string]string{"xact_commit": "1"},
	}, nil
}

func (c *fakeClient) getRollbacks() (*MetricStat, error) {
	return &MetricStat{
		database: c.database,
		stats:    map[string]string{"xact_rollback": "2"},
	}, nil
}

func (c *fakeClient) getBackends() (*MetricStat, error) {
	return &MetricStat{
		database: c.database,
		stats:    map[string]string{"count": "3"},
	}, nil
}

func (c *fakeClient) getDatabaseSize() (*MetricStat, error) {
	return &MetricStat{
		database: c.database,
		stats:    map[string]string{"db_size": "4"},
	}, nil
}

func (c *fakeClient) getDatabaseRowsByTable() ([]*MetricStat, error) {
	metricStats := []*MetricStat{}
	metricStats = append(metricStats, &MetricStat{
		database: c.database,
		table:    "public.table1",
		stats:    map[string]string{"live": "7", "dead": "8"},
	})
	metricStats = append(metricStats, &MetricStat{
		database: c.database,
		table:    "public.table2",
		stats:    map[string]string{"live": "9", "dead": "10"},
	})
	return metricStats, nil
}

func (c *fakeClient) getBlocksReadByTable() ([]*MetricStat, error) {
	metricStats := []*MetricStat{}
	metricStats = append(metricStats, &MetricStat{
		database: c.database,
		table:    "public.table1",
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
		table:    "public.table2",
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

func (c *fakeClient) getOperationsByTable() ([]*MetricStat, error) {
	metricStats := []*MetricStat{}
	metricStats = append(metricStats, &MetricStat{
		database: c.database,
		table:    "public.table1",
		stats: map[string]string{
			"ins":     "39",
			"upd":     "40",
			"del":     "41",
			"hot_upd": "42",
		},
	})
	metricStats = append(metricStats, &MetricStat{
		database: c.database,
		table:    "public.table2",
		stats: map[string]string{
			"ins":     "43",
			"upd":     "44",
			"del":     "45",
			"hot_upd": "46",
		},
	})
	return metricStats, nil
}
