package couchdbreceiver

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"

	"github.com/observiq/opentelemetry-components/receiver/couchdbreceiver/internal/metadata"
)

type couchdbScraper struct {
	logger *zap.Logger
	cfg    *Config
	client client
}

func newCouchdbScraper(logger *zap.Logger, cfg *Config) *couchdbScraper {
	return &couchdbScraper{
		logger: logger,
		cfg:    cfg,
	}
}

func (c *couchdbScraper) start(ctx context.Context, host component.Host) error {
	httpClient, err := newCouchDBClient(host, c.cfg, c.logger)
	if err != nil {
		c.logger.Error("failed to connect to couchdb", zap.Error(err))
		return err
	}
	c.client = httpClient
	return nil
}

// initMetric initializes a metric with a metadata label.
func initMetric(ms pdata.MetricSlice, mi metadata.MetricIntf) pdata.Metric {
	m := ms.AppendEmpty()
	mi.Init(m)
	return m
}

// addToDoubleMetric adds and labels a double gauge datapoint to a metricslice.
func addToDoubleMetric(metric pdata.NumberDataPointSlice, labels pdata.AttributeMap, value float64, ts pdata.Timestamp) {
	dataPoint := metric.AppendEmpty()
	dataPoint.SetTimestamp(ts)
	dataPoint.SetDoubleVal(value)
	if labels.Len() > 0 {
		labels.CopyTo(dataPoint.Attributes())
	}
}

// addToIntMetric adds and labels a int sum datapoint to metricslice.
func addToIntMetric(metric pdata.NumberDataPointSlice, labels pdata.AttributeMap, value int64, ts pdata.Timestamp) {
	dataPoint := metric.AppendEmpty()
	dataPoint.SetTimestamp(ts)
	dataPoint.SetIntVal(value)
	if labels.Len() > 0 {
		labels.CopyTo(dataPoint.Attributes())
	}
}

func (c *couchdbScraper) scrape(context.Context) (pdata.ResourceMetricsSlice, error) {
	if c.client == nil {
		return pdata.ResourceMetricsSlice{}, errors.New("failed to connect to couchdb client")
	}

	stats, err := c.client.Get()
	if err != nil {
		c.logger.Error("Failed to fetch couchdb metrics", zap.Error(err))
		return pdata.ResourceMetricsSlice{}, errors.New("failed to fetch couchdb stats")
	}

	// metric initialization
	rms := pdata.NewResourceMetricsSlice()
	ilm := rms.AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty()
	ilm.InstrumentationLibrary().SetName("otel/postgresql")
	now := pdata.NewTimestampFromTime(time.Now())

	requestTime := initMetric(ilm.Metrics(), metadata.M.CouchdbRequestTime).Gauge().DataPoints()
	httpdBulkRequests := initMetric(ilm.Metrics(), metadata.M.CouchdbHttpdBulkRequests).Sum().DataPoints()
	httpdRequestMethods := initMetric(ilm.Metrics(), metadata.M.CouchdbHttpdRequestMethods).Sum().DataPoints()
	httpdResponseCodes := initMetric(ilm.Metrics(), metadata.M.CouchdbHttpdResponseCodes).Sum().DataPoints()
	httpdTemporaryViewReads := initMetric(ilm.Metrics(), metadata.M.CouchdbHttpdTemporaryViewReads).Sum().DataPoints()
	viewReads := initMetric(ilm.Metrics(), metadata.M.CouchdbViewReads).Sum().DataPoints()
	openDatabases := initMetric(ilm.Metrics(), metadata.M.CouchdbOpenDatabases).Gauge().DataPoints()
	openFiles := initMetric(ilm.Metrics(), metadata.M.CouchdbOpenFiles).Gauge().DataPoints()
	reads := initMetric(ilm.Metrics(), metadata.M.CouchdbReads).Sum().DataPoints()
	writes := initMetric(ilm.Metrics(), metadata.M.CouchdbWrites).Sum().DataPoints()

	// request_time
	requestTimeLabels := pdata.NewAttributeMap()
	requestTimeLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	requestTimeKeys := []string{"request_time", "value", "arithmetic_mean"}
	requestTimeValue, err := getFloatFromBody(requestTimeKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "requestTime"),
		)
	} else {
		addToDoubleMetric(requestTime, requestTimeLabels, requestTimeValue, now)
	}

	// httpd bulk_requests
	httpdBulkRequestLabels := pdata.NewAttributeMap()
	httpdBulkRequestLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdBulkRequestKeys := []string{"httpd", "bulk_requests", "value"}
	httpdBulkRequestValue, err := getIntFromBody(httpdBulkRequestKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "bulkRequests"),
		)
	} else {
		addToIntMetric(httpdBulkRequests, httpdBulkRequestLabels, httpdBulkRequestValue, now)
	}

	// httpd_request_methods
	httpdRequestMethodsLabels := pdata.NewAttributeMap()
	httpdRequestMethodsLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdRequestMethodsLabels.Insert(metadata.L.HTTPMethod, pdata.NewAttributeValueString("COPY"))
	httpdRequestMethodsKeys := []string{"httpd_request_methods", "COPY", "value"}
	httpdRequestMethodsValue, err := getIntFromBody(httpdRequestMethodsKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdRequestMethods"),
		)
	} else {
		addToIntMetric(httpdRequestMethods, httpdRequestMethodsLabels, httpdRequestMethodsValue, now)
	}

	httpdRequestMethodsLabels = pdata.NewAttributeMap()
	httpdRequestMethodsLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdRequestMethodsLabels.Insert(metadata.L.HTTPMethod, pdata.NewAttributeValueString("DELETE"))
	httpdRequestMethodsKeys = []string{"httpd_request_methods", "DELETE", "value"}
	httpdRequestMethodsValue, err = getIntFromBody(httpdRequestMethodsKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdRequestMethods"),
		)
	} else {
		addToIntMetric(httpdRequestMethods, httpdRequestMethodsLabels, httpdRequestMethodsValue, now)
	}

	httpdRequestMethodsLabels = pdata.NewAttributeMap()
	httpdRequestMethodsLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdRequestMethodsLabels.Insert(metadata.L.HTTPMethod, pdata.NewAttributeValueString("GET"))
	httpdRequestMethodsKeys = []string{"httpd_request_methods", "GET", "value"}
	httpdRequestMethodsValue, err = getIntFromBody(httpdRequestMethodsKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdRequestMethods"),
		)
	} else {
		addToIntMetric(httpdRequestMethods, httpdRequestMethodsLabels, httpdRequestMethodsValue, now)
	}

	httpdRequestMethodsLabels = pdata.NewAttributeMap()
	httpdRequestMethodsLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdRequestMethodsLabels.Insert(metadata.L.HTTPMethod, pdata.NewAttributeValueString("HEAD"))
	httpdRequestMethodsKeys = []string{"httpd_request_methods", "HEAD", "value"}
	httpdRequestMethodsValue, err = getIntFromBody(httpdRequestMethodsKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdRequestMethods"),
		)
	} else {
		addToIntMetric(httpdRequestMethods, httpdRequestMethodsLabels, httpdRequestMethodsValue, now)
	}

	httpdRequestMethodsLabels = pdata.NewAttributeMap()
	httpdRequestMethodsLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdRequestMethodsLabels.Insert(metadata.L.HTTPMethod, pdata.NewAttributeValueString("OPTIONS"))
	httpdRequestMethodsKeys = []string{"httpd_request_methods", "OPTIONS", "value"}
	httpdRequestMethodsValue, err = getIntFromBody(httpdRequestMethodsKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdRequestMethods"),
		)
	} else {
		addToIntMetric(httpdRequestMethods, httpdRequestMethodsLabels, httpdRequestMethodsValue, now)
	}

	httpdRequestMethodsLabels = pdata.NewAttributeMap()
	httpdRequestMethodsLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdRequestMethodsKeys = []string{"httpd_request_methods", "POST", "value"}
	httpdRequestMethodsLabels.Insert(metadata.L.HTTPMethod, pdata.NewAttributeValueString("POST"))
	httpdRequestMethodsValue, err = getIntFromBody(httpdRequestMethodsKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdRequestMethods"),
		)
	} else {
		addToIntMetric(httpdRequestMethods, httpdRequestMethodsLabels, httpdRequestMethodsValue, now)
	}

	httpdRequestMethodsLabels = pdata.NewAttributeMap()
	httpdRequestMethodsLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdRequestMethodsLabels.Insert(metadata.L.HTTPMethod, pdata.NewAttributeValueString("PUT"))
	httpdRequestMethodsKeys = []string{"httpd_request_methods", "PUT", "value"}
	httpdRequestMethodsValue, err = getIntFromBody(httpdRequestMethodsKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdRequestMethods"),
		)
	} else {
		addToIntMetric(httpdRequestMethods, httpdRequestMethodsLabels, httpdRequestMethodsValue, now)
	}

	// httpd response codes
	httpdResponseCodesLabels := pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_200"))
	httpdResponseCodesKeys := []string{"httpd_status_codes", "200", "value"}
	httpdResponseCodesValue, err := getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_201"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "201", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_202"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "202", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_204"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "204", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_206"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "206", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_301"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "301", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_302"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "302", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_304"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "304", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_400"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "400", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_401"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "401", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_403"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "403", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_404"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "404", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_405"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "405", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_406"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "406", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_409"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "409", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_412"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "412", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_413"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "413", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_414"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "414", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_415"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "415", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_416"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "416", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_417"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "417", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_500"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "500", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_501"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "501", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	httpdResponseCodesLabels = pdata.NewAttributeMap()
	httpdResponseCodesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdResponseCodesLabels.Insert(metadata.L.ResponseCode, pdata.NewAttributeValueString("response_503"))
	httpdResponseCodesKeys = []string{"httpd_status_codes", "503", "value"}
	httpdResponseCodesValue, err = getIntFromBody(httpdResponseCodesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdResponseCodes"),
		)
	} else {
		addToIntMetric(httpdResponseCodes, httpdResponseCodesLabels, httpdResponseCodesValue, now)
	}

	// httpd temporary view reads
	httpdTemporaryViewReadsLabels := pdata.NewAttributeMap()
	httpdTemporaryViewReadsLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	httpdTemporaryViewReadsKeys := []string{"httpd", "temporary_view_reads", "value"}
	httpdTemporaryViewReadsValue, err := getIntFromBody(httpdTemporaryViewReadsKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "httpdTemporaryViewReads"),
		)
	} else {
		addToIntMetric(httpdTemporaryViewReads, httpdTemporaryViewReadsLabels, httpdTemporaryViewReadsValue, now)
	}

	// view reads
	viewReadsLabels := pdata.NewAttributeMap()
	viewReadsLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	viewReadsKeys := []string{"httpd", "view_reads", "value"}
	viewReadsValue, err := getIntFromBody(viewReadsKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "viewReads"),
		)
	} else {
		addToIntMetric(viewReads, viewReadsLabels, viewReadsValue, now)
	}

	// open databases
	openDatabasesLabels := pdata.NewAttributeMap()
	openDatabasesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	openDatabasesKeys := []string{"open_databases", "value"}
	openDatabasesValue, err := getFloatFromBody(openDatabasesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "openDatabases"),
		)
	} else {
		addToDoubleMetric(openDatabases, openDatabasesLabels, openDatabasesValue, now)
	}

	// open files
	openFilesLabels := pdata.NewAttributeMap()
	openFilesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	openFilesKeys := []string{"open_os_files", "value"}
	openFilesValue, err := getFloatFromBody(openFilesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "openFiles"),
		)
	} else {
		addToDoubleMetric(openFiles, openFilesLabels, openFilesValue, now)
	}

	// reads
	readsLabels := pdata.NewAttributeMap()
	readsLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	readsKeys := []string{"database_reads", "value"}
	readsValue, err := getIntFromBody(readsKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "reads"),
		)
	} else {
		addToIntMetric(reads, readsLabels, readsValue, now)
	}

	// writes
	writesLabels := pdata.NewAttributeMap()
	writesLabels.Insert(metadata.L.NodeName, pdata.NewAttributeValueString(c.cfg.Nodename))
	writesKeys := []string{"database_writes", "value"}
	writesValue, err := getIntFromBody(writesKeys, stats)
	if err != nil {
		c.logger.Info(
			err.Error(),
			zap.String("metric", "writes"),
		)
	} else {
		addToIntMetric(writes, writesLabels, writesValue, now)
	}

	return rms, nil
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
	case int64:
		return i, true
	case float64:
		return int64(i), true
	case float32:
		return int64(i), true
	case int32:
		return int64(i), true
	case string:
		intConv, err := strconv.ParseInt(i, 10, 64)
		if err != nil {
			return 0, false
		}
		return intConv, true
	}
	return 0, false
}

func getIntFromBody(keys []string, body map[string]interface{}) (int64, error) {
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
	intVal, ok := parseInt(currentValue)
	if !ok {
		return 0, fmt.Errorf("could not parse value as int")
	}
	return intVal, nil
}
