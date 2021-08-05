package elasticsearchreceiver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/observiq/opentelemetry-components/receiver/elasticsearchreceiver/internal/metadata"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"
)

type elasticsearchScraper struct {
	httpClient *http.Client
	logger     *zap.Logger
	cfg        *Config
	now        pdata.Timestamp
}

func newElasticSearchScraper(
	logger *zap.Logger,
	cfg *Config,
) (*elasticsearchScraper, error) {
	return &elasticsearchScraper{
		logger: logger,
		cfg:    cfg,
		now:    pdata.TimestampFromTime(time.Now()),
	}, nil
}

func (r *elasticsearchScraper) start(_ context.Context, host component.Host) error {
	httpClient, err := r.cfg.ToClient(host.GetExtensions())
	if err != nil {
		return err
	}
	r.httpClient = httpClient
	return nil
}

func (r *elasticsearchScraper) processFloatMetric(keys []string, body map[string]interface{}, metric pdata.NumberDataPointSlice, labels pdata.StringMap) {
	floatVal, err := getFloatFromBody(keys, body)
	if err != nil {
		r.logger.Info(err.Error())
	} else {
		addToMetric(metric, labels, floatVal, r.now)
	}
}

func (r *elasticsearchScraper) processIntMetric(keys []string, body map[string]interface{}, metric pdata.NumberDataPointSlice, labels pdata.StringMap) {
	intVal, err := getIntFromBody(keys, body)
	if err != nil {
		r.logger.Info(err.Error())
	} else {
		addToIntMetric(metric, labels, intVal, r.now)
	}
}

func (r *elasticsearchScraper) makeRequest(path string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", r.cfg.Endpoint+path, nil)
	if err != nil {
		return nil, err
	}
	if r.cfg.Username != "" && r.cfg.Password != "" {
		req.Header.Add("Authorization", "Basic "+basicAuth(r.cfg.Username, r.cfg.Password))
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var bodyParsed map[string]interface{}
	err = json.Unmarshal(body, &bodyParsed)
	if err != nil {
		return nil, err
	}
	return bodyParsed, nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (r *elasticsearchScraper) scrape(context.Context) (pdata.ResourceMetricsSlice, error) {
	nodeStats, err := r.makeRequest("/_nodes/stats")
	if err != nil {
		return pdata.ResourceMetricsSlice{}, err
	}

	r.now = pdata.TimestampFromTime(time.Now())
	metrics := pdata.NewMetrics()
	ilm := metrics.ResourceMetrics().AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty()
	ilm.InstrumentationLibrary().SetName("otel/elasticsearch")
	nodesInter := nodeStats["nodes"]
	nodes, ok := nodesInter.(map[string]interface{})
	if !ok {
		return pdata.ResourceMetricsSlice{}, fmt.Errorf("could not reflect set of nodes as a map")
	}

	cacheMemoryUsageMetric := initMetric(ilm.Metrics(), metadata.M.ElasticsearchCacheMemoryUsage).Gauge().DataPoints()
	evictionsMetric := initMetric(ilm.Metrics(), metadata.M.ElasticsearchEvictions).Sum().DataPoints()
	GCCollectionsMetric := initMetric(ilm.Metrics(), metadata.M.ElasticsearchGcCollection).Sum().DataPoints()
	MemoryUsageMetric := initMetric(ilm.Metrics(), metadata.M.ElasticsearchMemoryUsage).Gauge().DataPoints()
	NetworkMetric := initMetric(ilm.Metrics(), metadata.M.ElasticsearchNetwork).Sum().DataPoints()
	CurrentDocsMetric := initMetric(ilm.Metrics(), metadata.M.ElasticsearchCurrentDocuments).Gauge().DataPoints()
	DataNodesMetric := initMetric(ilm.Metrics(), metadata.M.ElasticsearchDataNodes).Gauge().DataPoints()
	HTTPConnsMetric := initMetric(ilm.Metrics(), metadata.M.ElasticsearchHTTPConnections).Gauge().DataPoints()
	NodesMetric := initMetric(ilm.Metrics(), metadata.M.ElasticsearchNodes).Gauge().DataPoints()
	OpenFilesMetric := initMetric(ilm.Metrics(), metadata.M.ElasticsearchOpenFiles).Gauge().DataPoints()
	ServerConnsMetric := initMetric(ilm.Metrics(), metadata.M.ElasticsearchServerConnections).Gauge().DataPoints()
	ShardsMetric := initMetric(ilm.Metrics(), metadata.M.ElasticsearchShards).Gauge().DataPoints()
	OperationsMetric := initMetric(ilm.Metrics(), metadata.M.ElasticsearchOperations).Sum().DataPoints()
	OperationTimeMetric := initMetric(ilm.Metrics(), metadata.M.ElasticsearchOperationTime).Sum().DataPoints()
	peakThreadsMetric := initMetric(ilm.Metrics(), metadata.M.ElasticsearchPeakThreads).Gauge().DataPoints()
	storageSizeMetric := initMetric(ilm.Metrics(), metadata.M.ElasticsearchStorageSize).Gauge().DataPoints()
	threadsMetric := initMetric(ilm.Metrics(), metadata.M.ElasticsearchThreads).Gauge().DataPoints()

	for nodeName, nodeDataInter := range nodes {
		labels := pdata.NewStringMap()
		labels.Upsert(metadata.L.ServerName, nodeName)

		nodeData, ok := nodeDataInter.(map[string]interface{})
		if !ok {
			r.logger.Info("could not reflect node data as a map")
			continue
		}
		labels.Upsert(metadata.L.CacheName, "query")
		r.processFloatMetric([]string{"indices", "query_cache", "memory_size_in_bytes"}, nodeData, cacheMemoryUsageMetric, labels)
		labels.Upsert(metadata.L.CacheName, "request")
		r.processFloatMetric([]string{"indices", "request_cache", "memory_size_in_bytes"}, nodeData, cacheMemoryUsageMetric, labels)
		labels.Upsert(metadata.L.CacheName, "field")
		r.processFloatMetric([]string{"indices", "fielddata", "memory_size_in_bytes"}, nodeData, cacheMemoryUsageMetric, labels)

		labels.Upsert(metadata.L.CacheName, "query")
		r.processIntMetric([]string{"indices", "query_cache", "evictions"}, nodeData, evictionsMetric, labels)
		labels.Upsert(metadata.L.CacheName, "request")
		r.processIntMetric([]string{"indices", "request_cache", "evictions"}, nodeData, evictionsMetric, labels)
		labels.Upsert(metadata.L.CacheName, "field")
		r.processIntMetric([]string{"indices", "fielddata", "evictions"}, nodeData, evictionsMetric, labels)
		labels.Delete(metadata.L.CacheName)

		labels.Upsert(metadata.L.GcType, "young")
		r.processIntMetric([]string{"jvm", "gc", "collectors", "young", "collection_count"}, nodeData, GCCollectionsMetric, labels)
		labels.Upsert(metadata.L.GcType, "old")
		r.processIntMetric([]string{"jvm", "gc", "collectors", "old", "collection_count"}, nodeData, GCCollectionsMetric, labels)
		labels.Delete(metadata.L.GcType)

		labels.Upsert(metadata.L.MemoryType, "heap")
		r.processFloatMetric([]string{"jvm", "mem", "heap_used_in_bytes"}, nodeData, MemoryUsageMetric, labels)
		labels.Upsert(metadata.L.MemoryType, "non-heap")
		r.processFloatMetric([]string{"jvm", "mem", "non_heap_used_in_bytes"}, nodeData, MemoryUsageMetric, labels)
		labels.Delete(metadata.L.MemoryType)

		labels.Upsert(metadata.L.Direction, "receive")
		r.processIntMetric([]string{"transport", "rx_size_in_bytes"}, nodeData, NetworkMetric, labels)
		labels.Upsert(metadata.L.Direction, "transmit")
		r.processIntMetric([]string{"transport", "tx_size_in_bytes"}, nodeData, NetworkMetric, labels)
		labels.Delete(metadata.L.Direction)

		labels.Upsert(metadata.L.DocumentType, "live")
		r.processFloatMetric([]string{"indices", "docs", "count"}, nodeData, CurrentDocsMetric, labels)
		labels.Upsert(metadata.L.DocumentType, "deleted")
		r.processFloatMetric([]string{"indices", "docs", "deleted"}, nodeData, CurrentDocsMetric, labels)
		labels.Delete(metadata.L.DocumentType)

		r.processFloatMetric([]string{"http", "current_open"}, nodeData, HTTPConnsMetric, labels)

		r.processFloatMetric([]string{"process", "open_file_descriptors"}, nodeData, OpenFilesMetric, labels)

		r.processFloatMetric([]string{"transport", "server_open"}, nodeData, ServerConnsMetric, labels)

		labels.Upsert(metadata.L.Operation, "index")
		r.processIntMetric([]string{"indices", "indexing", "index_total"}, nodeData, OperationsMetric, labels)
		labels.Upsert(metadata.L.Operation, "delete")
		r.processIntMetric([]string{"indices", "indexing", "delete_total"}, nodeData, OperationsMetric, labels)
		labels.Upsert(metadata.L.Operation, "get")
		r.processIntMetric([]string{"indices", "get", "total"}, nodeData, OperationsMetric, labels)
		labels.Upsert(metadata.L.Operation, "query")
		r.processIntMetric([]string{"indices", "search", "query_total"}, nodeData, OperationsMetric, labels)
		labels.Upsert(metadata.L.Operation, "fetch")
		r.processIntMetric([]string{"indices", "search", "fetch_total"}, nodeData, OperationsMetric, labels)

		labels.Upsert(metadata.L.Operation, "index")
		r.processIntMetric([]string{"indices", "indexing", "index_time_in_millis"}, nodeData, OperationTimeMetric, labels)
		labels.Upsert(metadata.L.Operation, "delete")
		r.processIntMetric([]string{"indices", "indexing", "delete_time_in_millis"}, nodeData, OperationTimeMetric, labels)
		labels.Upsert(metadata.L.Operation, "get")
		r.processIntMetric([]string{"indices", "get", "time_in_millis"}, nodeData, OperationTimeMetric, labels)
		labels.Upsert(metadata.L.Operation, "query")
		r.processIntMetric([]string{"indices", "search", "query_time_in_millis"}, nodeData, OperationTimeMetric, labels)
		labels.Upsert(metadata.L.Operation, "fetch")
		r.processIntMetric([]string{"indices", "search", "fetch_time_in_millis"}, nodeData, OperationTimeMetric, labels)
		labels.Delete(metadata.L.Operation)

		r.processFloatMetric([]string{"jvm", "threads", "peak_count"}, nodeData, peakThreadsMetric, labels)

		r.processFloatMetric([]string{"indices", "store", "size_in_bytes"}, nodeData, storageSizeMetric, labels)

		r.processFloatMetric([]string{"jvm", "threads", "count"}, nodeData, threadsMetric, labels)
	}

	labels := pdata.NewStringMap()

	clusterStats, err := r.makeRequest("/_cluster/stats")
	if err != nil {
		return pdata.ResourceMetricsSlice{}, err
	}
	r.processFloatMetric([]string{"nodes", "count", "data"}, clusterStats, DataNodesMetric, labels)

	r.processFloatMetric([]string{"nodes", "count", "total"}, clusterStats, NodesMetric, labels)

	clusterHealth, err := r.makeRequest("/_cluster/health")
	if err != nil {
		return pdata.ResourceMetricsSlice{}, err
	}
	labels.Upsert(metadata.L.ShardType, "initializing")
	r.processFloatMetric([]string{"initializing_shards"}, clusterHealth, ShardsMetric, labels)
	labels.Upsert(metadata.L.ShardType, "relocating")
	r.processFloatMetric([]string{"relocating_shards"}, clusterHealth, ShardsMetric, labels)
	labels.Upsert(metadata.L.ShardType, "active")
	r.processFloatMetric([]string{"active_shards"}, clusterHealth, ShardsMetric, labels)
	labels.Upsert(metadata.L.ShardType, "unassigned")
	r.processFloatMetric([]string{"unassigned_shards"}, clusterHealth, ShardsMetric, labels)
	labels.Delete(metadata.L.ShardType)

	return metrics.ResourceMetrics(), nil
}

func getFloatFromBody(keys []string, body map[string]interface{}) (float64, error) {
	var currentValue interface{} = body

	for _, key := range keys {
		currentBody, ok := currentValue.(map[string]interface{})
		if !ok {
			return 0, fmt.Errorf("could not find key in body")
		}

		currentValue, ok = currentBody[key]
		if !ok {
			return 0, fmt.Errorf("could not find key in body")
		}
	}
	floatVal, ok := parseFloat(currentValue)
	if !ok {
		return 0, fmt.Errorf("could not parse value as float")
	}
	return floatVal, nil
}

func getIntFromBody(keys []string, body map[string]interface{}) (int64, error) {
	var currentValue interface{} = body

	for _, key := range keys {
		currentBody, ok := currentValue.(map[string]interface{})
		if !ok {
			return 0, fmt.Errorf("could not find key in body, keys: %s", keys)
		}

		currentValue, ok = currentBody[key]
		if !ok {
			return 0, fmt.Errorf("could not find key in body, keys: %s", keys)
		}
	}
	intVal, ok := parseInt(currentValue)
	if !ok {
		return 0, fmt.Errorf("could not parse value as int, keys: %s", keys)
	}
	return intVal, nil
}

// parseFloat converts string to float64.
func parseFloat(value interface{}) (float64, bool) {
	switch f := value.(type) {
	case float64:
		return f, true
	case int64:
		return float64(f), true
	case float32:
		return float64(f), true
	case int32:
		return float64(f), true
	case string:
		fConv, err := strconv.ParseFloat(f, 64)
		if err != nil {
			return 0, false
		}
		return fConv, true
	}
	return 0, false
}

func parseInt(value interface{}) (int64, bool) {
	switch i := value.(type) {
	case float64:
		return int64(i), true
	case int64:
		return i, true
	case float32:
		return int64(i), true
	case int32:
		return int64(i), true
	case string:
		iConv, err := strconv.ParseInt(i, 10, 64)
		if err != nil {
			return 0, false
		}
		return iConv, true
	}
	return 0, false
}

func initMetric(ms pdata.MetricSlice, mi metadata.MetricIntf) pdata.Metric {
	m := ms.AppendEmpty()
	mi.Init(m)
	return m
}

func addToMetric(metric pdata.NumberDataPointSlice, labels pdata.StringMap, value float64, ts pdata.Timestamp) {
	dataPoint := metric.AppendEmpty()
	dataPoint.SetTimestamp(ts)
	dataPoint.SetDoubleVal(value)
	if labels.Len() > 0 {
		labels.CopyTo(dataPoint.LabelsMap())
	}
}

func addToIntMetric(metric pdata.NumberDataPointSlice, labels pdata.StringMap, value int64, ts pdata.Timestamp) {
	dataPoint := metric.AppendEmpty()
	dataPoint.SetTimestamp(ts)
	dataPoint.SetIntVal(value)
	if labels.Len() > 0 {
		labels.CopyTo(dataPoint.LabelsMap())
	}
}
