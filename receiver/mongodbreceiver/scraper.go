package mongodbreceiver

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/zap"

	"github.com/observiq/opentelemetry-components/receiver/mongodbreceiver/internal/metadata"
)

type mongodbScraper struct {
	logger *zap.Logger
	config *Config
}

type numberType int

const (
	integer numberType = iota
	double
)

type mongoMetric struct {
	metricDef     metadata.MetricIntf
	path          []string
	staticLabels  map[string]string
	dataPointType numberType
}

var dbStatsMetrics []mongoMetric = []mongoMetric{
	mongoMetric{
		metricDef:     metadata.M.MongodbCollections,
		path:          []string{"collections"},
		dataPointType: integer,
	},
	mongoMetric{
		metricDef:     metadata.M.MongodbDataSize,
		path:          []string{"dataSize"},
		dataPointType: double,
	},
	mongoMetric{
		metricDef:     metadata.M.MongodbExtents,
		path:          []string{"numExtents"},
		dataPointType: integer,
	},
	mongoMetric{
		metricDef:     metadata.M.MongodbIndexSize,
		path:          []string{"indexSize"},
		dataPointType: double,
	},
	mongoMetric{
		metricDef:     metadata.M.MongodbIndexes,
		path:          []string{"indexes"},
		dataPointType: integer,
	},
	mongoMetric{
		metricDef:     metadata.M.MongodbObjects,
		path:          []string{"objects"},
		dataPointType: integer,
	},
	mongoMetric{
		metricDef:     metadata.M.MongodbStorageSize,
		path:          []string{"storageSize"},
		dataPointType: double,
	},
}

var serverStatusMetrics []mongoMetric = []mongoMetric{
	mongoMetric{
		metricDef:     metadata.M.MongodbConnections,
		path:          []string{"connections", "current"},
		dataPointType: integer,
	},
	mongoMetric{
		metricDef:     metadata.M.MongodbMemoryUsage,
		path:          []string{"mem", "resident"},
		staticLabels:  map[string]string{metadata.L.MemoryType: metadata.LabelMemoryType.Resident},
		dataPointType: integer,
	},
	mongoMetric{
		metricDef:     metadata.M.MongodbMemoryUsage,
		path:          []string{"mem", "virtual"},
		staticLabels:  map[string]string{metadata.L.MemoryType: metadata.LabelMemoryType.Virtual},
		dataPointType: integer,
	},
	mongoMetric{
		metricDef:     metadata.M.MongodbMemoryUsage,
		path:          []string{"mem", "mapped"},
		staticLabels:  map[string]string{metadata.L.MemoryType: metadata.LabelMemoryType.Mapped},
		dataPointType: integer,
	},
	mongoMetric{
		metricDef:     metadata.M.MongodbMemoryUsage,
		path:          []string{"mem", "mappedWithJournal"},
		staticLabels:  map[string]string{metadata.L.MemoryType: metadata.LabelMemoryType.MappedWithJournal},
		dataPointType: integer,
	},
}

func newMongodbScraper(
	logger *zap.Logger,
	config *Config,
) scraperhelper.Scraper {
	ms := &mongodbScraper{
		logger: logger,
		config: config,
	}
	return scraperhelper.NewResourceMetricsScraper(config.ID(), ms.scrape)
}

func (r *mongodbScraper) scrape(ctx context.Context) (pdata.ResourceMetricsSlice, error) {
	// Init client in scrape method in case there are transient errors in the
	// constructor.
	client, err := r.initClient(ctx, r.logger, r.config.Timeout)
	if err != nil {
		r.logger.Error("Failed to connect to mongodb", zap.Error(err))
		return pdata.ResourceMetricsSlice{}, err
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			r.logger.Error("Failed to disconnect from client", zap.Error(err))
		}
	}()

	metrics := pdata.NewMetrics()
	ilm := metrics.ResourceMetrics().AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty()
	ilm.InstrumentationLibrary().SetName("otelcol/mongodb")
	mm := newMetricManager(r.logger, ilm)

	timeoutCtx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()
	dbNames, err := client.ListDatabaseNames(timeoutCtx, bson.D{})
	if err != nil {
		r.logger.Error("fetch database names", zap.Error(err))
		return pdata.ResourceMetricsSlice{}, err
	}

	serverStatus, err := client.query(ctx, "admin", bson.M{"serverStatus": 1})
	if err != nil {
		r.logger.Error("query serverStatus in admin", zap.Error(err))
	} else {
		r.parseSpecialMetrics(ctx, mm, serverStatus)
	}

	for _, dbName := range dbNames {
		dbStats, err := client.query(ctx, dbName, bson.M{"dbStats": 1})
		if err != nil {
			r.logger.Error("collect dbStats metric", zap.Error(err), zap.String("database", dbName))
		} else {
			r.parseDatabaseMetrics(ctx, mm, dbName, dbStatsMetrics, dbStats)
		}

		serverStatus, err := client.query(ctx, dbName, bson.M{"serverStatus": 1})
		if err != nil {
			r.logger.Error("collect serverStatus metric", zap.Error(err), zap.String("database", dbName))
		} else {
			r.parseDatabaseMetrics(ctx, mm, dbName, serverStatusMetrics, serverStatus)
		}
	}

	return metrics.ResourceMetrics(), nil
}

func (r *mongodbScraper) parseSpecialMetrics(ctx context.Context, mm *metricManager, document bson.M) {
	// Collect Global Lock Wait Time

	// Mongo version older than 3.0
	waitTime, err := getIntMetricValue(document, []string{"globalLock", "totalTime"})
	if err == nil {
		mm.addIntDataPoint(metadata.M.MongodbGlobalLockHoldTime, waitTime, pdata.NewStringMap())
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
			mm.addIntDataPoint(metadata.M.MongodbGlobalLockHoldTime, totalWaitTime, pdata.NewStringMap())
		}
	}

	// Collect Cache Hits & Misses
	canCalculateCacheHits := true

	cacheMisses, err := getIntMetricValue(document, []string{"wiredTiger", "cache", "pages read into cache"})
	if err != nil {
		r.logger.Error("parsing: ", zap.Error(err), zap.String("metric", metadata.M.MongodbCacheMisses.Name()))
		canCalculateCacheHits = false
	} else {
		mm.addIntDataPoint(metadata.M.MongodbCacheMisses, cacheMisses, pdata.NewStringMap())
	}

	totalCacheRequests, err := getIntMetricValue(document, []string{"wiredTiger", "cache", "pages requested from the cache"})
	if err != nil {
		r.logger.Error("parsing: ", zap.Error(err), zap.String("metric", metadata.M.MongodbCacheHits.Name()))
		canCalculateCacheHits = false
	}

	if canCalculateCacheHits && totalCacheRequests > cacheMisses {
		cacheHits := totalCacheRequests - cacheMisses
		mm.addIntDataPoint(metadata.M.MongodbCacheHits, cacheHits, pdata.NewStringMap())
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
			r.logger.Error("parsing: ", zap.Error(err), zap.String("metric", metadata.M.MongodbOperationCount.Name()))
		} else {
			labels := pdata.NewStringMap()
			labels.Insert(metadata.L.Operation, operation)
			mm.addIntDataPoint(metadata.M.MongodbOperationCount, count, labels)
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
		labels := pdata.NewStringMap()
		labels.Insert(metadata.L.DatabaseName, databaseName)
		for k, v := range metricRequest.staticLabels {
			labels.Insert(k, v)
		}

		switch metricRequest.dataPointType {
		case integer:
			value, err := getIntMetricValue(document, metricRequest.path)
			if err != nil {
				r.logger.Error("parsing: ", zap.Error(err), zap.String("metric", metricRequest.metricDef.Name()))
				continue
			}
			mm.addIntDataPoint(metricRequest.metricDef, value, labels)
		case double:
			value, err := getDoubleMetricValue(document, metricRequest.path)
			if err != nil {
				r.logger.Error("parsing: ", zap.Error(err), zap.String("metric", metricRequest.metricDef.Name()))
				continue
			}
			mm.addDoubleDataPoint(metricRequest.metricDef, value, labels)
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
	now                pdata.Timestamp
}

func newMetricManager(logger *zap.Logger, ilm pdata.InstrumentationLibraryMetrics) *metricManager {
	return &metricManager{
		logger:             logger,
		ilm:                ilm,
		initializedMetrics: map[string]pdata.Metric{},
		now:                pdata.TimestampFromTime(time.Now()),
	}
}

func (m *metricManager) addIntDataPoint(metricDef metadata.MetricIntf, value int64, labels pdata.StringMap) {
	dataPoints := m.getOrInit(metricDef)
	dataPoint := dataPoints.AppendEmpty()
	dataPoint.SetTimestamp(m.now)
	dataPoint.SetIntVal(value)
	labels.CopyTo(dataPoint.LabelsMap())
}

func (m *metricManager) addDoubleDataPoint(metricDef metadata.MetricIntf, value float64, labels pdata.StringMap) {
	dataPoints := m.getOrInit(metricDef)
	dataPoint := dataPoints.AppendEmpty()
	dataPoint.SetTimestamp(m.now)
	dataPoint.SetDoubleVal(value)
	labels.CopyTo(dataPoint.LabelsMap())
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

	m.logger.Error("unknown type", zap.String("metric", metricDef.Name()))
	return pdata.NewNumberDataPointSlice()
}
