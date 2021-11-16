package mongodbreceiver

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"

	"github.com/observiq/opentelemetry-components/receiver/mongodbreceiver/internal/metadata"
)

type mongodbScraper struct {
	logger *zap.Logger
	config *Config
	client Client
}

type numberType int

const (
	integer numberType = iota
	double
)

type mongoMetric struct {
	metricDef        metadata.MetricIntf
	path             []string
	staticAttributes map[string]string
	dataPointType    numberType
}

var dbStatsMetrics = []mongoMetric{
	{
		metricDef:     metadata.M.MongodbCollections,
		path:          []string{"collections"},
		dataPointType: integer,
	},
	{
		metricDef:     metadata.M.MongodbDataSize,
		path:          []string{"dataSize"},
		dataPointType: double,
	},
	{
		metricDef:     metadata.M.MongodbExtents,
		path:          []string{"numExtents"},
		dataPointType: integer,
	},
	{
		metricDef:     metadata.M.MongodbIndexSize,
		path:          []string{"indexSize"},
		dataPointType: double,
	},
	{
		metricDef:     metadata.M.MongodbIndexes,
		path:          []string{"indexes"},
		dataPointType: integer,
	},
	{
		metricDef:     metadata.M.MongodbObjects,
		path:          []string{"objects"},
		dataPointType: integer,
	},
	{
		metricDef:     metadata.M.MongodbStorageSize,
		path:          []string{"storageSize"},
		dataPointType: double,
	},
}

var serverStatusMetrics = []mongoMetric{
	{
		metricDef:        metadata.M.MongodbConnections,
		path:             []string{"connections", "active"},
		staticAttributes: map[string]string{metadata.L.ConnectionType: metadata.LabelConnectionType.Active},
		dataPointType:    integer,
	},
	{
		metricDef:        metadata.M.MongodbConnections,
		path:             []string{"connections", "available"},
		staticAttributes: map[string]string{metadata.L.ConnectionType: metadata.LabelConnectionType.Available},
		dataPointType:    integer,
	},
	{
		metricDef:        metadata.M.MongodbConnections,
		path:             []string{"connections", "current"},
		staticAttributes: map[string]string{metadata.L.ConnectionType: metadata.LabelConnectionType.Current},
		dataPointType:    integer,
	},
	{
		metricDef:        metadata.M.MongodbMemoryUsage,
		path:             []string{"mem", "resident"},
		staticAttributes: map[string]string{metadata.L.MemoryType: metadata.LabelMemoryType.Resident},
		dataPointType:    integer,
	},
	{
		metricDef:        metadata.M.MongodbMemoryUsage,
		path:             []string{"mem", "virtual"},
		staticAttributes: map[string]string{metadata.L.MemoryType: metadata.LabelMemoryType.Virtual},
		dataPointType:    integer,
	},
	{
		metricDef:        metadata.M.MongodbMemoryUsage,
		path:             []string{"mem", "mapped"},
		staticAttributes: map[string]string{metadata.L.MemoryType: metadata.LabelMemoryType.Mapped},
		dataPointType:    integer,
	},
	{
		metricDef:        metadata.M.MongodbMemoryUsage,
		path:             []string{"mem", "mappedWithJournal"},
		staticAttributes: map[string]string{metadata.L.MemoryType: metadata.LabelMemoryType.MappedWithJournal},
		dataPointType:    integer,
	},
}

func newMongodbScraper(logger *zap.Logger, config *Config) (*mongodbScraper, error) {
	ms := &mongodbScraper{
		logger: logger,
		config: config,
	}
	return ms, nil
}

func (r *mongodbScraper) start(ctx context.Context, host component.Host) error {
	clientLogger := r.logger.Named("mongo-client")
	client := NewClient(r.config, clientLogger)
	r.client = client
	return nil
}

func (r *mongodbScraper) shutdown(ctx context.Context) error {
	if r.client != nil {
		return r.client.Disconnect(ctx)
	}
	return nil
}

func (r *mongodbScraper) scrape(ctx context.Context) (pdata.ResourceMetricsSlice, error) {
	if r.client == nil {
		return pdata.NewResourceMetricsSlice(), errors.New("no client was initialized before calling scrape")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	if err := r.client.Connect(timeoutCtx); err != nil {
		r.logger.Error("Failed to connect to client", zap.Error(err))
		return pdata.NewResourceMetricsSlice(), err
	}

	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := r.client.Ping(ctx, readpref.PrimaryPreferred())
	if err != nil {
		r.logger.Error("Failed to ping server", zap.Error(err))
		return pdata.NewResourceMetricsSlice(), err
	}

	return r.collectMetrics(ctx, r.client)
}

func (r *mongodbScraper) collectMetrics(ctx context.Context, client Client) (pdata.ResourceMetricsSlice, error) {
	rms := pdata.NewResourceMetricsSlice()
	ilm := rms.AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty()
	ilm.InstrumentationLibrary().SetName("otelcol/mongodb")
	mm := newMetricManager(r.logger, ilm)

	timeoutCtx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()
	dbNames, err := client.ListDatabaseNames(timeoutCtx, bson.D{})
	if err != nil {
		r.logger.Error("Failed to fetch database names", zap.Error(err))
		return pdata.ResourceMetricsSlice{}, err
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go r.collectAdminDatabase(ctx, wg, mm, client)

	for _, dbName := range dbNames {
		wg.Add(1)
		go r.collectDatabase(ctx, wg, mm, client, dbName)
	}

	wg.Wait()

	return rms, nil
}

func (r *mongodbScraper) collectDatabase(ctx context.Context, wg *sync.WaitGroup, mm *metricManager, client Client, databaseName string) {
	defer wg.Done()
	dbStats, err := client.Query(ctx, databaseName, bson.M{"dbStats": 1})
	if err != nil {
		r.logger.Error("Failed to collect dbStats metric", zap.Error(err), zap.String("database", databaseName))
	} else {
		r.parseDatabaseMetrics(ctx, mm, databaseName, dbStatsMetrics, dbStats)
	}

	serverStatus, err := client.Query(ctx, databaseName, bson.M{"serverStatus": 1})
	if err != nil {
		r.logger.Error("Failed to collect serverStatus metric", zap.Error(err), zap.String("database", databaseName))
	} else {
		r.parseDatabaseMetrics(ctx, mm, databaseName, serverStatusMetrics, serverStatus)
	}
}

func (r *mongodbScraper) collectAdminDatabase(ctx context.Context, wg *sync.WaitGroup, mm *metricManager, client Client) {
	defer wg.Done()
	serverStatus, err := client.Query(ctx, "admin", bson.M{"serverStatus": 1})
	if err != nil {
		r.logger.Error("Failed to query serverStatus in admin", zap.Error(err))
	} else {
		r.parseSpecialMetrics(ctx, mm, serverStatus)
	}
}

func (r *mongodbScraper) parseSpecialMetrics(ctx context.Context, mm *metricManager, document bson.M) {
	// Collect Global Lock Wait Time

	// Mongo version older than 3.0
	waitTime, err := getIntMetricValue(document, []string{"globalLock", "totalTime"})
	if err == nil {
		mm.addDataPoint(metadata.M.MongodbGlobalLockHoldTime, waitTime, pdata.NewAttributeMap())
	} else {
		// Assume mongoDB 3.0+ if the older style was not available
		totalWaitTime := int64(0)
		hasValue := false
		for _, lockType := range []string{"W", "R", "r", "w"} {
			waitTime, err := getIntMetricValue(document, []string{"locks", "Global", "timeAcquiringMicros", lockType})
			if err == nil {
				totalWaitTime += int64(waitTime / 1000)
				hasValue = true
			}
		}

		if hasValue {
			mm.addDataPoint(metadata.M.MongodbGlobalLockHoldTime, totalWaitTime, pdata.NewAttributeMap())
		}
	}

	// Collect Cache Hits & Misses
	canCalculateCacheHits := true

	cacheMisses, err := getIntMetricValue(document, []string{"wiredTiger", "cache", "pages read into cache"})
	if err != nil {
		r.logger.Error("Failed to Parse", zap.Error(err), zap.String("metric", metadata.M.MongodbCacheMisses.Name()))
		canCalculateCacheHits = false
	} else {
		mm.addDataPoint(metadata.M.MongodbCacheMisses, cacheMisses, pdata.NewAttributeMap())
	}

	totalCacheRequests, err := getIntMetricValue(document, []string{"wiredTiger", "cache", "pages requested from the cache"})
	if err != nil {
		r.logger.Error("Failed to Parse", zap.Error(err), zap.String("metric", metadata.M.MongodbCacheHits.Name()))
		canCalculateCacheHits = false
	}

	if canCalculateCacheHits && totalCacheRequests > cacheMisses {
		cacheHits := totalCacheRequests - cacheMisses
		mm.addDataPoint(metadata.M.MongodbCacheHits, cacheHits, pdata.NewAttributeMap())
	}

	// Collect Operations
	for _, operation := range []string{
		metadata.LabelOperation.Insert,
		metadata.LabelOperation.Query,
		metadata.LabelOperation.Update,
		metadata.LabelOperation.Delete,
		metadata.LabelOperation.Getmore,
		metadata.LabelOperation.Command,
	} {
		count, err := getIntMetricValue(document, []string{"opcounters", operation})
		if err != nil {
			r.logger.Error("Failed to Parse", zap.Error(err), zap.String("metric", metadata.M.MongodbOperations.Name()))
		} else {
			attributes := pdata.NewAttributeMap()
			attributes.Insert(metadata.L.Operation, pdata.NewAttributeValueString(operation))
			mm.addDataPoint(metadata.M.MongodbOperations, count, attributes)
		}
	}
}

func (r *mongodbScraper) parseDatabaseMetrics(
	ctx context.Context,
	mm *metricManager,
	databaseName string,
	metricsRequested []mongoMetric,
	document bson.M,
) {
	for _, metricRequest := range metricsRequested {
		attributes := pdata.NewAttributeMap()
		attributes.Insert(metadata.L.DatabaseName, pdata.NewAttributeValueString(databaseName))
		for k, v := range metricRequest.staticAttributes {
			attributes.Insert(k, pdata.NewAttributeValueString(v))
		}

		switch metricRequest.dataPointType {
		case integer:
			value, err := getIntMetricValue(document, metricRequest.path)
			if err != nil {
				r.logger.Error("Failed to Parse", zap.Error(err), zap.String("metric", metricRequest.metricDef.Name()))
				continue
			}
			mm.addDataPoint(metricRequest.metricDef, value, attributes)
		case double:
			value, err := getDoubleMetricValue(document, metricRequest.path)
			if err != nil {
				r.logger.Error("Failed to Parse", zap.Error(err), zap.String("metric", metricRequest.metricDef.Name()))
				continue
			}
			mm.addDataPoint(metricRequest.metricDef, value, attributes)
		}
	}
}

func getIntMetricValue(document bson.M, path []string) (int64, error) {
	curItem, remainingPath := path[0], path[1:]
	value := document[curItem]
	if value == nil {
		return 0, errors.New("nil found when digging for metric")
	} else if len(remainingPath) == 0 {
		switch v := value.(type) {
		case int:
			return int64(v), nil
		case int32:
			return int64(v), nil
		case int64:
			return v, nil
		case string:
			return strconv.ParseInt(v, 10, 64)
		default:
			return 0, fmt.Errorf("unexpected type found when parsing int: %v", reflect.TypeOf(value))
		}
	} else {
		return getIntMetricValue(value.(bson.M), remainingPath)
	}
}

func getDoubleMetricValue(document bson.M, path []string) (float64, error) {
	curItem, remainingPath := path[0], path[1:]
	value := document[curItem]
	if value == nil {
		return 0, errors.New("nil found when digging for metric")
	} else if len(remainingPath) == 0 {
		switch v := value.(type) {
		case float32:
			return float64(v), nil
		case float64:
			return v, nil
		case string:
			return strconv.ParseFloat(v, 64)
		default:
			return 0, fmt.Errorf("unexpected type found when parsing double: %v", reflect.TypeOf(value))
		}
	} else {
		return getDoubleMetricValue(value.(bson.M), remainingPath)
	}
}

type metricManager struct {
	logger             *zap.Logger
	ilm                pdata.InstrumentationLibraryMetrics
	initializedMetrics map[string]pdata.Metric
	mutex              *sync.RWMutex
	now                pdata.Timestamp
}

func newMetricManager(logger *zap.Logger, ilm pdata.InstrumentationLibraryMetrics) *metricManager {
	mutex := &sync.RWMutex{}
	return &metricManager{
		logger:             logger,
		ilm:                ilm,
		initializedMetrics: map[string]pdata.Metric{},
		mutex:              mutex,
		now:                pdata.NewTimestampFromTime(time.Now()),
	}
}

func (m *metricManager) addDataPoint(metricDef metadata.MetricIntf, value interface{}, attributes pdata.AttributeMap) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	currDatapoints := m.getOrInit(metricDef)
	dataPoint := currDatapoints.AppendEmpty()
	dataPoint.SetTimestamp(m.now)
	switch v := value.(type) {
	case int64:
		dataPoint.SetIntVal(v)
	case float64:
		dataPoint.SetDoubleVal(v)
	default:
		m.logger.Warn(fmt.Sprintf("unknown metric data type for metric: %s", metricDef.Name()))
		return
	}
	attributes.CopyTo(dataPoint.Attributes())
}

func (m *metricManager) getOrInit(metricDef metadata.MetricIntf) pdata.NumberDataPointSlice {
	metric, ok := m.initializedMetrics[metricDef.Name()]
	if !ok {
		metric = m.ilm.Metrics().AppendEmpty()
		metricDef.Init(metric)
		m.initializedMetrics[metricDef.Name()] = metric
	}

	if metric.DataType() == pdata.MetricDataTypeSum {
		return metric.Sum().DataPoints()
	}

	if metric.DataType() == pdata.MetricDataTypeGauge {
		return metric.Gauge().DataPoints()
	}

	m.logger.Error("Failed to get or init metric of unknown type", zap.String("metric", metricDef.Name()))
	return pdata.NewNumberDataPointSlice()
}
