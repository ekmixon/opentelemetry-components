package couchbasereceiver

type Metric struct {
	Name       string
	Timestamp  int
	Value      string
	ErrMessage string
}

var Metrics = []Metric{
	{
		Name: "kv_curr_connections",
	},
	{
		Name: "kv_collection_data_size_bytes",
	},
	{
		Name: "kv_collection_item_count",
	},
	{
		Name: "kv_collection_mem_used_bytes",
	},
	{
		Name: "kv_collection_ops",
	},
	{
		Name: "kv_curr_connections",
	},
	{
		Name: "kv_curr_items",
	},
	{
		Name: "kv_curr_items_tot",
	},
	// {
	// 	Name: "kv_item_pager_seconds_bucket",
	// },
	// {
	// 	Name: "kv_item_pager_seconds_count",
	// },
	// {
	// 	Name: "kv_item_pager_seconds_sum",
	// },
	{
		Name: "kv_ops",
	},
	{
		Name: "kv_ops_failed",
	},
	{
		Name: "kv_read_bytes",
	},
	{
		Name: "kv_system_connections",
	},
	{
		Name: "kv_time_seconds",
	},
	{
		Name: "kv_total_connections",
	},
	{
		Name: "kv_total_memory_used_bytes",
	},
	{
		Name: "kv_total_resp_errors",
	},
	{
		Name: "kv_uptime_seconds",
	},
	{
		Name: "kv_written_bytes",
	},
}
