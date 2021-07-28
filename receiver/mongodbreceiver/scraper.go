package mongodbreceiver

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/zap"

	"github.com/observiq/opentelemetry-components/receiver/mongodbreceiver/internal/metadata"
)

type mongodbScraper struct {
	logger *zap.Logger
	config *Config
}

type mongoMetric struct {
	metricDef        metadata.MetricIntf
	path             []string
	additionalLabels pdata.StringMap
}

var dbStatsMetrics []mongoMetric = []mongoMetric{
	mongoMetric{
		metricDef:        metadata.M.MongodbCollections,
		path:             []string{"collections"},
		additionalLabels: pdata.NewStringMap(),
	},
	mongoMetric{
		metricDef:        metadata.M.MongodbDataSize,
		path:             []string{"dataSize"},
		additionalLabels: pdata.NewStringMap(),
	},
	mongoMetric{
		metricDef:        metadata.M.MongodbExtents,
		path:             []string{"numExtents"},
		additionalLabels: pdata.NewStringMap(),
	},
	mongoMetric{
		metricDef:        metadata.M.MongodbIndexSize,
		path:             []string{"indexSize"},
		additionalLabels: pdata.NewStringMap(),
	},
	mongoMetric{
		metricDef:        metadata.M.MongodbIndexes,
		path:             []string{"indexes"},
		additionalLabels: pdata.NewStringMap(),
	},
	mongoMetric{
		metricDef:        metadata.M.MongodbObjects,
		path:             []string{"objects"},
		additionalLabels: pdata.NewStringMap(),
	},
	mongoMetric{
		metricDef:        metadata.M.MongodbStorageSize,
		path:             []string{"storageSize"},
		additionalLabels: pdata.NewStringMap(),
	},
}

var serverStatusMetrics []mongoMetric = []mongoMetric{
	mongoMetric{
		metricDef:        metadata.M.MongodbConnections,
		path:             []string{"connections", "current"},
		additionalLabels: pdata.NewStringMap(),
	},
	mongoMetric{
		metricDef:        metadata.M.MongodbMemoryUsage,
		path:             []string{"mem", "resident"},
		additionalLabels: pdata.NewStringMap().InitFromMap(map[string]string{metadata.L.MemoryType: metadata.LabelMemoryType.Resident}),
	},
	mongoMetric{
		metricDef:        metadata.M.MongodbMemoryUsage,
		path:             []string{"mem", "virtual"},
		additionalLabels: pdata.NewStringMap().InitFromMap(map[string]string{metadata.L.MemoryType: metadata.LabelMemoryType.Virtual}),
	},
	mongoMetric{
		metricDef:        metadata.M.MongodbMemoryUsage,
		path:             []string{"mem", "mapped"},
		additionalLabels: pdata.NewStringMap().InitFromMap(map[string]string{metadata.L.MemoryType: metadata.LabelMemoryType.Mapped}),
	},
	mongoMetric{
		metricDef:        metadata.M.MongodbMemoryUsage,
		path:             []string{"mem", "mappedWithJournal"},
		additionalLabels: pdata.NewStringMap().InitFromMap(map[string]string{metadata.L.MemoryType: metadata.LabelMemoryType.MappedWithJournal}),
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
	client, err := r.initClient(ctx)
	if err != nil {
		r.logger.Error("Failed to connect to mongodb", zap.Error(err))
		return pdata.ResourceMetricsSlice{}, err
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			r.logger.Error("Failed to disconnect from client", zap.Error(err))
		}
	}()

	now := pdata.TimestampFromTime(time.Now())
	metrics := pdata.NewMetrics()
	ilm := metrics.ResourceMetrics().AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty()
	ilm.InstrumentationLibrary().SetName("otelcol/mongodb")

	timeoutCtx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()
	databaseNames, err := client.ListDatabaseNames(timeoutCtx, bson.D{})
	if err != nil {
		r.logger.Error("Failed to fetch mongodb database names", zap.Error(err))
		return pdata.ResourceMetricsSlice{}, err
	}

	err = r.collectSpecialMetrics(ctx, client, ilm, now)
	if err != nil {
		r.logger.Error("Failed to collect mongoDB metrics from serverStatus in admin", zap.Error(err))
	}

	initializedMetrics := map[string]pdata.Metric{}
	for _, databaseName := range databaseNames {
		err = r.runDatabaseCommandAndCollectMetrics(ctx, client, databaseName, now, ilm, initializedMetrics, dbStatsMetrics, true, bson.M{"dbStats": 1})
		if err != nil {
			r.logger.Error(fmt.Sprintf("Failed to collect mongoDB metrics from dbStats in database %s", databaseName), zap.Error(err))
		}

		err = r.runDatabaseCommandAndCollectMetrics(ctx, client, databaseName, now, ilm, initializedMetrics, serverStatusMetrics, true, bson.M{"serverStatus": 1})
		if err != nil {
			r.logger.Error(fmt.Sprintf("Failed to collect mongoDB metrics from serverStatus in database %s", databaseName), zap.Error(err))
		}
	}

	return metrics.ResourceMetrics(), nil
}

func getOrInitializeMetric(initializedMetrics map[string]pdata.Metric, ilm pdata.InstrumentationLibraryMetrics, requestedMetric mongoMetric) pdata.Metric {
	value, ok := initializedMetrics[requestedMetric.metricDef.Name()]
	if !ok {
		metric := ilm.Metrics().AppendEmpty()
		requestedMetric.metricDef.Init(metric)
		initializedMetrics[requestedMetric.metricDef.Name()] = metric
		return metric
	}

	return value
}

func (r *mongodbScraper) collectSpecialMetrics(
	ctx context.Context,
	client *mongo.Client,
	ilm pdata.InstrumentationLibraryMetrics,
	now pdata.Timestamp,
) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()
	result := client.Database("admin").RunCommand(timeoutCtx, bson.M{"serverStatus": 1})

	var document bson.M
	err := result.Decode(&document)

	if err != nil {
		return err
	}

	// Collect Global Lock Wait Time

	// Mongo version older than 3.0
	path := []string{"globalLock", "totalTime"}
	waitTime, err := getIntMetricValue(document, path)
	if err == nil {
		lockTimeMetric := ilm.Metrics().AppendEmpty()
		metadata.M.MongodbGlobalLockHoldTime.Init(lockTimeMetric)
		addIntDataPoint(lockTimeMetric.IntSum().DataPoints(), now, waitTime, pdata.NewStringMap())
	} else {
		// Assume mongoDB 3.0+ if the older style was not available
		totalWaitTime := int64(0)
		hasValue := false
		for _, lockType := range []string{"W", "R", "r", "w"} {
			path = []string{"locks", "Global", "timeAcquiringMicros", lockType}
			waitTime, err := getIntMetricValue(document, path)
			if err == nil {
				totalWaitTime += int64(waitTime / 1000)
				hasValue = true
			}
		}

		if hasValue {
			lockTimeMetric := ilm.Metrics().AppendEmpty()
			metadata.M.MongodbGlobalLockHoldTime.Init(lockTimeMetric)
			addIntDataPoint(lockTimeMetric.IntSum().DataPoints(), now, totalWaitTime, pdata.NewStringMap())
		}
	}

	// Collect Cache Hits & Misses
	canCalculateCacheHits := true

	cacheMisses, err := getIntMetricValue(document, []string{"wiredTiger", "cache", "pages read into cache"})
	if err != nil {
		r.logger.Error("Error while collecting mongodb cache misses metric", zap.Error(err))
		canCalculateCacheHits = false
	} else {
		cacheMissesMetric := ilm.Metrics().AppendEmpty()
		metadata.M.MongodbCacheMisses.Init(cacheMissesMetric)
		addIntDataPoint(cacheMissesMetric.IntSum().DataPoints(), now, cacheMisses, pdata.NewStringMap())
	}

	totalCacheRequests, err := getIntMetricValue(document, []string{"wiredTiger", "cache", "pages requested from the cache"})
	if err != nil {
		r.logger.Error("Error while collecting mongodb cache requests metric", zap.Error(err))
		canCalculateCacheHits = false
	}

	if canCalculateCacheHits && totalCacheRequests > cacheMisses {
		cacheHitsMetric := ilm.Metrics().AppendEmpty()
		metadata.M.MongodbCacheHits.Init(cacheHitsMetric)
		cacheHits := totalCacheRequests - cacheMisses
		addIntDataPoint(cacheHitsMetric.IntSum().DataPoints(), now, cacheHits, pdata.NewStringMap())
	}

	// Collect Operations
	var operationsMetric *pdata.Metric
	for _, operation := range []string{
		metadata.LabelOperation.Insert,
		metadata.LabelOperation.Query,
		metadata.LabelOperation.Update,
		metadata.LabelOperation.Delete,
		metadata.LabelOperation.Getmore,
		metadata.LabelOperation.Command,
	} {
		path := []string{"opcounters", operation}
		count, err := getIntMetricValue(document, path)
		if err != nil {
			r.logger.Error("Error while collecting mongodb operations metric", zap.Error(err))
		} else {
			if operationsMetric == nil {
				metric := ilm.Metrics().AppendEmpty()
				operationsMetric = &metric
				metadata.M.MongodbOperationCount.Init(*operationsMetric)
			}

			addIntDataPoint(operationsMetric.IntSum().DataPoints(), now, count, pdata.NewStringMap().InitFromMap(map[string]string{metadata.L.Operation: operation}))
		}
	}

	return nil
}

func (r *mongodbScraper) runDatabaseCommandAndCollectMetrics(
	ctx context.Context,
	client *mongo.Client,
	databaseName string,
	now pdata.Timestamp,
	ilm pdata.InstrumentationLibraryMetrics,
	initializedMetrics map[string]pdata.Metric,
	metricsRequested []mongoMetric,
	includeDatabaseNameLabel bool,
	command bson.M,
) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()
	result := client.Database(databaseName).RunCommand(timeoutCtx, command)

	var document bson.M
	err := result.Decode(&document)
	if err != nil {
		r.logger.Error("Error while collecting mongodb metric", zap.Error(err))
		return err
	}

	for _, metricRequest := range metricsRequested {
		metric := getOrInitializeMetric(initializedMetrics, ilm, metricRequest)
		labels := pdata.NewStringMap()
		metricRequest.additionalLabels.CopyTo(labels)
		if includeDatabaseNameLabel {
			labels.Insert(metadata.L.DatabaseName, databaseName)
		}

		switch metric.DataType() {
		case pdata.MetricDataTypeIntGauge:
			value, err := getIntMetricValue(document, metricRequest.path)
			if err == nil {
				addIntDataPoint(metric.IntGauge().DataPoints(), now, value, labels)
			} else {
				r.logger.Error("Failed to collect mongodb metric", zap.Error(err))
			}
		case pdata.MetricDataTypeIntSum:
			value, err := getIntMetricValue(document, metricRequest.path)
			if err == nil {
				addIntDataPoint(metric.IntSum().DataPoints(), now, value, labels)
			} else {
				r.logger.Error("Failed to collect mongodb metric", zap.Error(err))
			}
		case pdata.MetricDataTypeSum:
			value, err := getDoubleMetricValue(document, metricRequest.path)
			if err == nil {
				addDoubleDataPoint(metric.Sum().DataPoints(), now, value, labels)
			} else {
				r.logger.Error("Failed to collect mongodb metric", zap.Error(err))
			}
		case pdata.MetricDataTypeGauge:
			value, err := getDoubleMetricValue(document, metricRequest.path)
			if err == nil {
				addDoubleDataPoint(metric.Gauge().DataPoints(), now, value, labels)
			} else {
				r.logger.Error("Failed to collect mongodb metric", zap.Error(err))
			}
		}
	}

	return nil
}

func (r *mongodbScraper) initClient(ctx context.Context) (*mongo.Client, error) {
	authentication := ""
	if r.config.Username != "" && r.config.Password != "" {
		authentication = fmt.Sprintf("%s:%s@", r.config.Username, r.config.Password)
	}

	uri := fmt.Sprintf("mongodb://%s%s", authentication, r.config.Endpoint)

	timeoutCtx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	return mongo.Connect(timeoutCtx, options.Client().ApplyURI(uri))
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

func addIntDataPoint(slice pdata.IntDataPointSlice, now pdata.Timestamp, value int64, labels pdata.StringMap) {
	dp := slice.AppendEmpty()
	dp.SetTimestamp(now)
	dp.SetValue(value)
	labels.CopyTo(dp.LabelsMap())
}

func addDoubleDataPoint(slice pdata.DoubleDataPointSlice, now pdata.Timestamp, value float64, labels pdata.StringMap) {
	dp := slice.AppendEmpty()
	dp.SetTimestamp(now)
	dp.SetValue(value)
	labels.CopyTo(dp.LabelsMap())
}
