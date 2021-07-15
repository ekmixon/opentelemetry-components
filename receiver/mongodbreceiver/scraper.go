// Copyright 2020, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mongodbreceiver

import (
	"context"
	"errors"
	"fmt"
	"reflect"
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
	metricDef metadata.MetricIntf
	path      []string
	dataType  pdata.MetricDataType
}

var dbStatsMetrics []mongoMetric = []mongoMetric{
	mongoMetric{
		metricDef: metadata.M.MongodbCollections,
		path:      []string{"collections"},
		dataType:  pdata.MetricDataTypeIntGauge,
	},
	mongoMetric{
		metricDef: metadata.M.MongodbDataSize,
		path:      []string{"dataSize"},
		dataType:  pdata.MetricDataTypeGauge,
	},
	mongoMetric{
		metricDef: metadata.M.MongodbExtents,
		path:      []string{"numExtents"},
		dataType:  pdata.MetricDataTypeIntGauge,
	},
	mongoMetric{
		metricDef: metadata.M.MongodbIndexSize,
		path:      []string{"indexSize"},
		dataType:  pdata.MetricDataTypeGauge,
	},
	mongoMetric{
		metricDef: metadata.M.MongodbIndexes,
		path:      []string{"indexes"},
		dataType:  pdata.MetricDataTypeIntGauge,
	},
	mongoMetric{
		metricDef: metadata.M.MongodbObjects,
		path:      []string{"objects"},
		dataType:  pdata.MetricDataTypeIntGauge,
	},
	mongoMetric{
		metricDef: metadata.M.MongodbStorageSize,
		path:      []string{"storageSize"},
		dataType:  pdata.MetricDataTypeGauge,
	},
}

// MongodbCacheHits          MetricIntf
// MongodbCacheMisses        MetricIntf
// MongodbGlobalLockHoldTime MetricIntf

// MongodbOperationCount     MetricIntf

// DB Stats
// MongodbCollections        MetricIntf
// MongodbDataSize           MetricIntf
// MongodbIndexSize          MetricIntf
// MongodbIndexes            MetricIntf
// MongodbObjects            MetricIntf
// MongodbStorageSize        MetricIntf
// MongodbExtents            MetricIntf

// MongodbConnections        MetricIntf
// MongodbMemoryUsage        MetricIntf

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

	initializedMetrics := map[string]pdata.Metric{}
	for _, dbStatsMetric := range dbStatsMetrics {
		metric := ilm.Metrics().AppendEmpty()
		dbStatsMetric.metricDef.Init(metric)
		initializedMetrics[dbStatsMetric.metricDef.Name()] = metric
	}

	for _, databaseName := range databaseNames {
		timeoutCtx, cancel = context.WithTimeout(ctx, r.config.Timeout)
		defer cancel()
		result := client.Database(databaseName).RunCommand(timeoutCtx, bson.M{"dbStats": 1})

		var document bson.M
		err = result.Decode(&document)

		if err != nil {
			r.logger.Error("Failed to fetch mongodb database stats", zap.Error(err))
			continue
		}

		for _, dbStatsMetric := range dbStatsMetrics {
			metric := initializedMetrics[dbStatsMetric.metricDef.Name()]
			labels := pdata.NewStringMap()
			labels.Insert(metadata.L.DatabaseName, databaseName)

			switch dbStatsMetric.dataType {
			case pdata.MetricDataTypeIntGauge:
				value, err := getIntMetricValue(document, dbStatsMetric.path)
				if err == nil {
					addIntDataPoint(metric.IntGauge().DataPoints(), now, value, labels)
				} else {
					r.logger.Error("Failed to collect mongodb metric", zap.Error(err))
				}
			case pdata.MetricDataTypeIntSum:
				value, err := getIntMetricValue(document, dbStatsMetric.path)
				if err == nil {
					addIntDataPoint(metric.IntSum().DataPoints(), now, value, labels)
				} else {
					r.logger.Error("Failed to collect mongodb metric", zap.Error(err))
				}
			case pdata.MetricDataTypeSum:
				value, err := getDoubleMetricValue(document, dbStatsMetric.path)
				if err == nil {
					addDoubleDataPoint(metric.Sum().DataPoints(), now, value, labels)
				} else {
					r.logger.Error("Failed to collect mongodb metric", zap.Error(err))
				}
			case pdata.MetricDataTypeGauge:
				value, err := getDoubleMetricValue(document, dbStatsMetric.path)
				if err == nil {
					addDoubleDataPoint(metric.Gauge().DataPoints(), now, value, labels)
				} else {
					r.logger.Error("Failed to collect mongodb metric", zap.Error(err))
				}
			}
		}
	}

	return metrics.ResourceMetrics(), nil
}

func (r *mongodbScraper) initClient(ctx context.Context) (*mongo.Client, error) {
	authentication := ""
	if r.config.User != nil && r.config.Password == nil {
		return nil, errors.New("user provided without password")
	} else if r.config.User == nil && r.config.Password != nil {
		return nil, errors.New("password provided without user")
	} else if r.config.User != nil && r.config.Password != nil {
		authentication = fmt.Sprintf("%s:%s@", *r.config.User, *r.config.Password)
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
			return parseInt(v), nil
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
			return parseFloat(v), nil
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
