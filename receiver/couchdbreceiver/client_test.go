package couchdbreceiver

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configtls"
	"go.uber.org/zap"
)

func TestURLParse(t *testing.T) {
	var err error
	// Default Protocol: HTTP
	// Default Port: depends on tech, but HTTP:80, HTTPS:443

	protocols := []string{"", "http://", "https://", "HTTP://", "HTTPS://"}
	hostnames := []string{"", "localhost", "127.0.0.1", "123.123.123.123"}
	ports := []string{"", ":80", ":8080", ":443", ":8443", ":1234"}

	for _, protocol := range protocols {
		for _, hostname := range hostnames {
			for _, port := range ports {
				endpoint := fmt.Sprintf("%s%s%s", protocol, hostname, port)
				t.Run(endpoint, func(t *testing.T) {
					_, err := url.Parse(endpoint)
					if err != nil {
						endpoint = "http://" + endpoint
						_, err = url.Parse(endpoint)
					}
					assert.NoError(t, err)
					// if ok := assert.NoError(t, err); ok {
					// 	assert.NotEmpty(t, url.Scheme)
					// 	assert.NotEmpty(t, url.Host)
					// }
				})
			}
		}
	}

	_, err = url.Parse("")
	assert.NoError(t, err, "empty empty empty")
}

func TestNewCouchDBClient(t *testing.T) {
	t.Run("failed", func(t *testing.T) {
		couchdbClient, err := newCouchDBClient(componenttest.NewNopHost(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: "",
				TLSSetting: configtls.TLSClientSetting{
					TLSSetting: configtls.TLSSetting{
						CAFile: "/non/existent",
					},
				},
			}}, zap.NewNop())

		require.Nil(t, couchdbClient)
		require.NotNil(t, err)
	})
	t.Run("no error", func(t *testing.T) {
		couchdbMock := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.URL.Path == "/_node/_local/_stats/couchdb" {
				rw.WriteHeader(200)
				_, _ = rw.Write([]byte(``))
				return
			}
			rw.WriteHeader(404)
		}))

		couchdbClient, err := newCouchDBClient(componenttest.NewNopHost(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: couchdbMock.URL + "/_node/_local/_stats/couchdb",
			},
		}, zap.NewNop())
		require.Nil(t, err)
		require.NotNil(t, couchdbClient)
	})
}

func TestGet(t *testing.T) {
	couchdbMock := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/_node/_local/_stats/couchdb" {
			rw.WriteHeader(200)
			_, _ = rw.Write([]byte(`{"key": "value"}`))
			return
		}
		if req.URL.Path == "/invalid_endpoint" {
			rw.WriteHeader(404)
			return
		}
		if req.URL.Path == "/invalid_body" {
			rw.Header().Set("Content-Length", "1")
			return
		}
		if req.URL.Path == "/invalid_json" {
			rw.WriteHeader(200)
			_, _ = rw.Write([]byte(`{"}`))
			return
		}

		rw.WriteHeader(404)
	}))
	t.Run("invalid endpoint", func(t *testing.T) {
		couchdbClient, err := newCouchDBClient(componenttest.NewNopHost(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: couchdbMock.URL + "/invalid_endpoint",
			},
		}, zap.NewNop())
		require.Nil(t, err)
		require.NotNil(t, couchdbClient)

		result, err := couchdbClient.Get()
		require.NotNil(t, err)
		require.Nil(t, result)
	})
	t.Run("invalid body", func(t *testing.T) {
		couchdbClient, err := newCouchDBClient(componenttest.NewNopHost(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: couchdbMock.URL + "/invalid_body",
			},
		}, zap.NewNop())
		require.Nil(t, err)
		require.NotNil(t, couchdbClient)

		result, err := couchdbClient.Get()
		require.NotNil(t, err)
		require.Nil(t, result)
	})
	t.Run("invalid json", func(t *testing.T) {
		couchdbClient, err := newCouchDBClient(componenttest.NewNopHost(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: couchdbMock.URL + "/invalid_json",
			},
		}, zap.NewNop())
		require.Nil(t, err)
		require.NotNil(t, couchdbClient)

		result, err := couchdbClient.Get()
		require.NotNil(t, err)
		require.Nil(t, result)
	})
	t.Run("no error", func(t *testing.T) {
		couchdbClient, err := newCouchDBClient(componenttest.NewNopHost(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: couchdbMock.URL + "/_node/_local/_stats/couchdb",
			},
		}, zap.NewNop())
		require.Nil(t, err)
		require.NotNil(t, couchdbClient)

		result, err := couchdbClient.Get()
		require.Nil(t, err)
		require.NotNil(t, result)
	})
}
