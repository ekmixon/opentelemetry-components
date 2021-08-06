// +build integration

package couchdbreceiver

import (
	"context"
	"fmt"
	"net"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/model/pdata"

	"github.com/observiq/opentelemetry-components/receiver/couchdbreceiver/internal/metadata"
)

func TestCouchdbIntegration(t *testing.T) {
	t.Run("Running docker version 3.0.0 on port 5985", func(t *testing.T) {
		container := getContainer(t, containerRequest300)
		defer func() {
			require.NoError(t, container.Terminate(context.Background()))
		}()
		hostname, err := container.Host(context.Background())
		require.NoError(t, err)

		f := NewFactory()
		cfg := f.CreateDefaultConfig().(*Config)
		cfg.Endpoint = net.JoinHostPort(hostname, "5984")
		cfg.Username = "otel"
		cfg.Password = "otel"
		cfg.Nodename = "_local"

		consumer := new(consumertest.MetricsSink)
		settings := componenttest.NewNopReceiverCreateSettings()
		rcvr, err := f.CreateMetricsReceiver(context.Background(), settings, cfg, consumer)
		require.NoError(t, err, "failed creating metrics receiver")
		require.NoError(t, rcvr.Start(context.Background(), componenttest.NewNopHost()))
		require.Eventuallyf(t, func() bool {
			return len(consumer.AllMetrics()) > 0
		}, 2*time.Minute, 1*time.Second, "failed to receive more than 0 metrics")

		md := consumer.AllMetrics()[0]
		require.Equal(t, 1, md.ResourceMetrics().Len())
		ilms := md.ResourceMetrics().At(0).InstrumentationLibraryMetrics()
		require.Equal(t, 1, ilms.Len())
		metrics := ilms.At(0).Metrics()
		require.NoError(t, rcvr.Shutdown(context.Background()))

		validateIntegrationResult(t, metrics)
	})

	t.Run("Running docker version 3.1.1 on port 5984", func(t *testing.T) {
		container := getContainer(t, containerRequest311)
		defer func() {
			require.NoError(t, container.Terminate(context.Background()))
		}()
		hostname, err := container.Host(context.Background())
		require.NoError(t, err)

		f := NewFactory()
		cfg := f.CreateDefaultConfig().(*Config)
		cfg.Endpoint = net.JoinHostPort(hostname, "5984")
		cfg.Username = "otel"
		cfg.Password = "otel"
		cfg.Nodename = "_local"

		consumer := new(consumertest.MetricsSink)
		settings := componenttest.NewNopReceiverCreateSettings()
		rcvr, err := f.CreateMetricsReceiver(context.Background(), settings, cfg, consumer)
		require.NoError(t, err, "failed creating metrics receiver")
		require.NoError(t, rcvr.Start(context.Background(), componenttest.NewNopHost()))
		require.Eventuallyf(t, func() bool {
			return len(consumer.AllMetrics()) > 0
		}, 2*time.Minute, 1*time.Second, "failed to receive more than 0 metrics")

		md := consumer.AllMetrics()[0]
		require.Equal(t, 1, md.ResourceMetrics().Len())
		ilms := md.ResourceMetrics().At(0).InstrumentationLibraryMetrics()
		require.Equal(t, 1, ilms.Len())
		metrics := ilms.At(0).Metrics()
		require.NoError(t, rcvr.Shutdown(context.Background()))

		validateIntegrationResult(t, metrics)
	})
}

var (
	containerRequest300 = testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.couchdb.300",
		},
		ExposedPorts: []string{"5984:5984"}, // TODO: standardize how to do ports between local-targets
		WaitingFor: wait.ForListeningPort("5984").
			WithStartupTimeout(2 * time.Minute),
	}
	containerRequest311 = testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.couchdb.311",
		},
		ExposedPorts: []string{"5984:5984"},
		WaitingFor: wait.ForListeningPort("5984").
			WithStartupTimeout(2 * time.Minute),
	}
)

func getContainer(t *testing.T, req testcontainers.ContainerRequest) testcontainers.Container {
	require.NoError(t, req.Validate())
	container, err := testcontainers.GenericContainer(
		context.Background(),
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
	require.NoError(t, err)
	return container
}

func validateIntegrationResult(t *testing.T, metric pdata.MetricSlice) {
	require.Equal(t, len(metadata.M.Names()), metric.Len())

	for i := 0; i < metric.Len(); i++ {
		m := metric.At(i)
		switch m.Name() {
		case metadata.M.CouchdbRequestTime.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
		case metadata.M.CouchdbHttpdBulkRequests.Name():
			dps := m.Sum().DataPoints()
			require.Equal(t, 1, dps.Len())
		case metadata.M.CouchdbHttpdRequestMethods.Name():
			dps := m.Sum().DataPoints()
			require.Equal(t, 7, dps.Len())

			requestMethodMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				method, _ := dp.LabelsMap().Get(metadata.L.HTTPMethod)
				label := fmt.Sprintf("%s method:%s", m.Name(), method)
				requestMethodMetrics[label] = true
			}

			require.Equal(t, 7, len(requestMethodMetrics))
			require.Equal(t, map[string]bool{
				"couchdb.httpd_request_methods method:COPY":    true,
				"couchdb.httpd_request_methods method:DELETE":  true,
				"couchdb.httpd_request_methods method:GET":     true,
				"couchdb.httpd_request_methods method:HEAD":    true,
				"couchdb.httpd_request_methods method:OPTIONS": true,
				"couchdb.httpd_request_methods method:POST":    true,
				"couchdb.httpd_request_methods method:PUT":     true,
			},
				requestMethodMetrics)
		case metadata.M.CouchdbHttpdResponseCodes.Name():
			dps := m.Sum().DataPoints()
			require.Equal(t, 24, dps.Len())

			respondCodeMetrics := map[string]bool{}
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				code, _ := dp.LabelsMap().Get(metadata.L.ResponseCode)
				label := fmt.Sprintf("%s code:%s", m.Name(), code)
				respondCodeMetrics[label] = true
			}

			require.Equal(t, 24, len(respondCodeMetrics))
			require.Equal(t, map[string]bool{
				"couchdb.httpd_response_codes code:response_200": true,
				"couchdb.httpd_response_codes code:response_201": true,
				"couchdb.httpd_response_codes code:response_202": true,
				"couchdb.httpd_response_codes code:response_204": true,
				"couchdb.httpd_response_codes code:response_206": true,
				"couchdb.httpd_response_codes code:response_301": true,
				"couchdb.httpd_response_codes code:response_302": true,
				"couchdb.httpd_response_codes code:response_304": true,
				"couchdb.httpd_response_codes code:response_400": true,
				"couchdb.httpd_response_codes code:response_401": true,
				"couchdb.httpd_response_codes code:response_403": true,
				"couchdb.httpd_response_codes code:response_404": true,
				"couchdb.httpd_response_codes code:response_405": true,
				"couchdb.httpd_response_codes code:response_406": true,
				"couchdb.httpd_response_codes code:response_409": true,
				"couchdb.httpd_response_codes code:response_412": true,
				"couchdb.httpd_response_codes code:response_413": true,
				"couchdb.httpd_response_codes code:response_414": true,
				"couchdb.httpd_response_codes code:response_415": true,
				"couchdb.httpd_response_codes code:response_416": true,
				"couchdb.httpd_response_codes code:response_417": true,
				"couchdb.httpd_response_codes code:response_500": true,
				"couchdb.httpd_response_codes code:response_501": true,
				"couchdb.httpd_response_codes code:response_503": true,
			},
				respondCodeMetrics)
		case metadata.M.CouchdbHttpdTemporaryViewReads.Name():
			dps := m.Sum().DataPoints()
			require.Equal(t, 1, dps.Len())
		case metadata.M.CouchdbViewReads.Name():
			dps := m.Sum().DataPoints()
			require.Equal(t, 1, dps.Len())
		case metadata.M.CouchdbOpenDatabases.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
		case metadata.M.CouchdbOpenFiles.Name():
			dps := m.Gauge().DataPoints()
			require.Equal(t, 1, dps.Len())
		case metadata.M.CouchdbReads.Name():
			dps := m.Sum().DataPoints()
			require.Equal(t, 1, dps.Len())
		case metadata.M.CouchdbWrites.Name():
			dps := m.Sum().DataPoints()
			require.Equal(t, 1, dps.Len())
		default:
			require.Nil(t, m.Name(), fmt.Sprintf("metrics %s not expected", m.Name()))
		}
	}
}
