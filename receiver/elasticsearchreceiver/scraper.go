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
	rms := pdata.NewResourceMetricsSlice()
	ilm := rms.AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty()
	ilm.InstrumentationLibrary().SetName("otel/elasticsearch")
	nodesInter, ok := nodeStats["nodes"]
	if !ok {
		return pdata.ResourceMetricsSlice{}, fmt.Errorf("no nodes data available")
	}
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
	threadPoolThreadsMetric := initMetric(ilm.Metrics(), metadata.M.ElasticsearchThreadPoolThreads).Gauge().DataPoints()
	threadPoolQueueMetric := initMetric(ilm.Metrics(), metadata.M.ElasticsearchThreadPoolQueue).Gauge().DataPoints()
	threadPoolActiveMetric := initMetric(ilm.Metrics(), metadata.M.ElasticsearchThreadPoolActive).Gauge().DataPoints()
	threadPoolRejectedMetric := initMetric(ilm.Metrics(), metadata.M.ElasticsearchThreadPoolRejected).Sum().DataPoints()
	threadPoolCompletedMetric := initMetric(ilm.Metrics(), metadata.M.ElasticsearchThreadPoolCompleted).Sum().DataPoints()

	for _, nodeDataInter := range nodes {
		nodeData, ok := nodeDataInter.(map[string]interface{})
		if !ok {
			r.logger.Error("could not reflect node data as a map")
			continue
		}

		nodeName, err := getStringFromBody([]string{"name"}, nodeData)
		if err != nil {
			r.logger.Error(err.Error())
			continue
		}

		labels := pdata.NewStringMap()
		labels.Upsert(metadata.L.ServerName, nodeName)
		labels.Upsert(metadata.L.CacheName, "query")
		r.processIntMetric([]string{"indices", "query_cache", "memory_size_in_bytes"}, nodeData, cacheMemoryUsageMetric, labels)
		labels.Upsert(metadata.L.CacheName, "request")
		r.processIntMetric([]string{"indices", "request_cache", "memory_size_in_bytes"}, nodeData, cacheMemoryUsageMetric, labels)
		labels.Upsert(metadata.L.CacheName, "field")
		r.processIntMetric([]string{"indices", "fielddata", "memory_size_in_bytes"}, nodeData, cacheMemoryUsageMetric, labels)

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
		r.processIntMetric([]string{"jvm", "mem", "heap_used_in_bytes"}, nodeData, MemoryUsageMetric, labels)
		labels.Upsert(metadata.L.MemoryType, "non-heap")
		r.processIntMetric([]string{"jvm", "mem", "non_heap_used_in_bytes"}, nodeData, MemoryUsageMetric, labels)
		labels.Delete(metadata.L.MemoryType)

		labels.Upsert(metadata.L.Direction, "receive")
		r.processIntMetric([]string{"transport", "rx_size_in_bytes"}, nodeData, NetworkMetric, labels)
		labels.Upsert(metadata.L.Direction, "transmit")
		r.processIntMetric([]string{"transport", "tx_size_in_bytes"}, nodeData, NetworkMetric, labels)
		labels.Delete(metadata.L.Direction)

		labels.Upsert(metadata.L.DocumentType, "live")
		r.processIntMetric([]string{"indices", "docs", "count"}, nodeData, CurrentDocsMetric, labels)
		labels.Upsert(metadata.L.DocumentType, "deleted")
		r.processIntMetric([]string{"indices", "docs", "deleted"}, nodeData, CurrentDocsMetric, labels)
		labels.Delete(metadata.L.DocumentType)

		r.processIntMetric([]string{"http", "current_open"}, nodeData, HTTPConnsMetric, labels)

		r.processIntMetric([]string{"process", "open_file_descriptors"}, nodeData, OpenFilesMetric, labels)

		r.processIntMetric([]string{"transport", "server_open"}, nodeData, ServerConnsMetric, labels)

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

		r.processIntMetric([]string{"jvm", "threads", "peak_count"}, nodeData, peakThreadsMetric, labels)

		r.processIntMetric([]string{"indices", "store", "size_in_bytes"}, nodeData, storageSizeMetric, labels)

		r.processIntMetric([]string{"jvm", "threads", "count"}, nodeData, threadsMetric, labels)

		threadPools, ok := nodeData["thread_pool"]
		if !ok {
			r.logger.Error("no thread pool data available")
			continue
		}
		threadPoolsInter, ok := threadPools.(map[string]interface{})
		if !ok {
			r.logger.Error("could not reflect thread pools data as a map")
			continue
		}

		for threadPoolName, threadPoolInter := range threadPoolsInter {
			threadPool, ok := threadPoolInter.(map[string]interface{})
			if !ok {
				r.logger.Error("could not reflect thread pool data as a map")
				continue
			}
			labels.Upsert(metadata.L.ThreadPoolName, threadPoolName)
			r.processIntMetric([]string{"threads"}, threadPool, threadPoolThreadsMetric, labels)
			r.processIntMetric([]string{"queue"}, threadPool, threadPoolQueueMetric, labels)
			r.processIntMetric([]string{"active"}, threadPool, threadPoolActiveMetric, labels)
			r.processIntMetric([]string{"rejected"}, threadPool, threadPoolRejectedMetric, labels)
			r.processIntMetric([]string{"completed"}, threadPool, threadPoolCompletedMetric, labels)
			labels.Delete(metadata.L.ThreadPoolName)
		}
	}

	labels := pdata.NewStringMap()

	clusterStats, err := r.makeRequest("/_cluster/stats")
	if err != nil {
		return pdata.ResourceMetricsSlice{}, err
	}
	r.processIntMetric([]string{"nodes", "count", "data"}, clusterStats, DataNodesMetric, labels)

	r.processIntMetric([]string{"nodes", "count", "total"}, clusterStats, NodesMetric, labels)

	clusterHealth, err := r.makeRequest("/_cluster/health")
	if err != nil {
		return pdata.ResourceMetricsSlice{}, err
	}
	labels.Upsert(metadata.L.ShardType, "initializing")
	r.processIntMetric([]string{"initializing_shards"}, clusterHealth, ShardsMetric, labels)
	labels.Upsert(metadata.L.ShardType, "relocating")
	r.processIntMetric([]string{"relocating_shards"}, clusterHealth, ShardsMetric, labels)
	labels.Upsert(metadata.L.ShardType, "active")
	r.processIntMetric([]string{"active_shards"}, clusterHealth, ShardsMetric, labels)
	labels.Upsert(metadata.L.ShardType, "unassigned")
	r.processIntMetric([]string{"unassigned_shards"}, clusterHealth, ShardsMetric, labels)
	labels.Delete(metadata.L.ShardType)

	return rms, nil
}

func getStringFromBody(keys []string, body map[string]interface{}) (string, error) {
	var currentValue interface{} = body

	for _, key := range keys {
		currentBody, ok := currentValue.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("could not find key in body")
		}

		currentValue, ok = currentBody[key]
		if !ok {
			return "", fmt.Errorf("could not find key in body")
		}
	}
	stringVal, ok := currentValue.(string)
	if !ok {
		return "", fmt.Errorf("could not parse value as string")
	}
	return stringVal, nil
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

func addToIntMetric(metric pdata.NumberDataPointSlice, labels pdata.StringMap, value int64, ts pdata.Timestamp) {
	dataPoint := metric.AppendEmpty()
	dataPoint.SetTimestamp(ts)
	dataPoint.SetIntVal(value)
	if labels.Len() > 0 {
		labels.CopyTo(dataPoint.LabelsMap())
	}
}
