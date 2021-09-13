package couchbasereceiver

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configtls"
	"go.uber.org/zap"
)

func TestNewCouchbaseClient(t *testing.T) {
	t.Run("failed", func(t *testing.T) {
		couchbaseClient, err := newCouchbaseClient(componenttest.NewNopHost(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: "",
				TLSSetting: configtls.TLSClientSetting{
					TLSSetting: configtls.TLSSetting{
						CAFile: "/non/existent",
					},
				},
			}}, zap.NewNop())

		require.Nil(t, couchbaseClient)
		require.Error(t, err)
	})
	t.Run("no error", func(t *testing.T) {
		couchbaseMock := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.URL.Path == "/pools/default" {
				rw.WriteHeader(200)
				_, _ = rw.Write([]byte(``))
				return
			}
			rw.WriteHeader(404)
		}))

		couchbaseClient, err := newCouchbaseClient(componenttest.NewNopHost(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: couchbaseMock.URL + "/pools/default",
			},
		}, zap.NewNop())
		require.NoError(t, err)
		require.NotNil(t, couchbaseClient)
	})
}

func TestGetNodeStats(t *testing.T) {
	couchbaseMock := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/pools/default/invalid_endpoint" {
			rw.WriteHeader(404)
			return
		}
		if req.URL.Path == "/invalid_read/pools/default" {
			rw.Header().Set("Content-Length", "1")
			return
		}
		if req.URL.Path == "/invalid_json/pools/default" {
			rw.WriteHeader(200)
			return
		}
		if req.URL.Path == "/no_error/pools/default" {
			rw.WriteHeader(200)
			_, _ = rw.Write([]byte(`{}`))
			return
		}
	}))
	t.Run("error on NewRequest", func(t *testing.T) {
		couchbaseClient, err := newCouchbaseClient(componenttest.NewNopHost(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: couchbaseMock.URL + "request with space",
			},
		}, zap.NewNop())
		require.NoError(t, err)
		require.NotNil(t, couchbaseClient)

		result, err := couchbaseClient.getNodeStats()
		require.Error(t, err)
		require.Nil(t, result)
	})

	t.Run("error on Do", func(t *testing.T) {
		couchbaseClient, err := newCouchbaseClient(componenttest.NewNopHost(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: "",
			},
		}, zap.NewNop())
		require.NoError(t, err)
		require.NotNil(t, couchbaseClient)

		result, err := couchbaseClient.getNodeStats()
		require.Error(t, err)
		require.Nil(t, result)
	})
	t.Run("error on read", func(t *testing.T) {
		couchbaseClient, err := newCouchbaseClient(componenttest.NewNopHost(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: couchbaseMock.URL + "/invalid_read",
			},
		}, zap.NewNop())
		require.NoError(t, err)
		require.NotNil(t, couchbaseClient)

		result, err := couchbaseClient.getNodeStats()
		require.Error(t, err)
		require.Nil(t, result)
	})
	t.Run("error on unmarshal", func(t *testing.T) {
		couchbaseClient, err := newCouchbaseClient(componenttest.NewNopHost(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: couchbaseMock.URL + "/invalid_json",
			},
		}, zap.NewNop())
		require.NoError(t, err)
		require.NotNil(t, couchbaseClient)

		result, err := couchbaseClient.getNodeStats()
		require.Error(t, err)
		require.Nil(t, result)
	})
	t.Run("no error", func(t *testing.T) {
		couchbaseClient, err := newCouchbaseClient(componenttest.NewNopHost(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: couchbaseMock.URL + "/no_error",
			},
		}, zap.NewNop())
		require.NoError(t, err)
		require.NotNil(t, couchbaseClient)

		result, err := couchbaseClient.getNodeStats()
		require.NoError(t, err)
		require.NotNil(t, result)
	})
}

func TestGetBucketsStats(t *testing.T) {
	couchbaseMock := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/invalid_read" {
			rw.Header().Set("Content-Length", "1")
			return
		}
		if req.URL.Path == "/invalid_json" {
			rw.WriteHeader(200)
			return
		}
		if req.URL.Path == "/pools/default/buckets" {
			rw.WriteHeader(200)
			_, _ = rw.Write([]byte(`[{}]`))
			return
		}
	}))
	t.Run("error on NewRequest", func(t *testing.T) {
		couchbaseClient, err := newCouchbaseClient(componenttest.NewNopHost(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: couchbaseMock.URL,
			},
		}, zap.NewNop())
		require.NoError(t, err)
		require.NotNil(t, couchbaseClient)

		result, err := couchbaseClient.getBucketsStats("request with space")
		require.Error(t, err)
		require.Nil(t, result)
	})
	t.Run("error on Do", func(t *testing.T) {
		couchbaseClient, err := newCouchbaseClient(componenttest.NewNopHost(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: "",
			},
		}, zap.NewNop())
		require.NoError(t, err)
		require.NotNil(t, couchbaseClient)

		result, err := couchbaseClient.getBucketsStats("/invalid_request")
		require.Error(t, err)
		require.Nil(t, result)
	})
	t.Run("error on read", func(t *testing.T) {
		couchbaseClient, err := newCouchbaseClient(componenttest.NewNopHost(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: couchbaseMock.URL,
			},
		}, zap.NewNop())
		require.NoError(t, err)
		require.NotNil(t, couchbaseClient)

		result, err := couchbaseClient.getBucketsStats("/invalid_read")
		require.Error(t, err)
		require.Nil(t, result)
	})
	t.Run("error on unmarshal", func(t *testing.T) {
		couchbaseClient, err := newCouchbaseClient(componenttest.NewNopHost(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: couchbaseMock.URL,
			},
		}, zap.NewNop())
		require.NoError(t, err)
		require.NotNil(t, couchbaseClient)

		result, err := couchbaseClient.getBucketsStats("/invalid_json")
		require.Error(t, err)
		require.Nil(t, result)
	})
	t.Run("no error", func(t *testing.T) {
		couchbaseClient, err := newCouchbaseClient(componenttest.NewNopHost(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: couchbaseMock.URL,
			},
		}, zap.NewNop())
		require.NoError(t, err)
		require.NotNil(t, couchbaseClient)

		result, err := couchbaseClient.getBucketsStats("/pools/default/buckets")
		require.NoError(t, err)
		require.NotNil(t, result)
	})
}

func TestGet(t *testing.T) {
	couchbaseNodeStatsMock := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/invalid_json" {
			rw.WriteHeader(200)
			return
		}
		if req.URL.Path == "/error_on_buckets/pools/default" {
			rw.WriteHeader(200)
			_, _ = rw.Write([]byte(`{}`))
			return
		}
		if req.URL.Path == "/no_error/pools/default" {
			rw.WriteHeader(200)
			_, _ = rw.Write([]byte(`{
				"buckets": {"uri": "/pools/default/buckets?v=105443202&uuid=4517d7f94cde4b64cdd260a662c65072"}}`))
			return
		}
		if req.URL.Path == "/no_error/pools/default/buckets" {
			rw.WriteHeader(200)
			_, _ = rw.Write([]byte(`[{}]`))
			return
		}
	}))
	t.Run("error on nodestats", func(t *testing.T) {
		couchbaseClient, err := newCouchbaseClient(componenttest.NewNopHost(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: couchbaseNodeStatsMock.URL + "/invalid_json",
			},
		}, zap.NewNop())
		require.NoError(t, err)
		require.NotNil(t, couchbaseClient)

		result, err := couchbaseClient.Get()
		require.Error(t, err)
		require.Nil(t, result)
	})
	t.Run("error on bucketsStats", func(t *testing.T) {
		couchbaseClient, err := newCouchbaseClient(componenttest.NewNopHost(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: couchbaseNodeStatsMock.URL + "/error_on_buckets",
			},
		}, zap.NewNop())
		require.NoError(t, err)
		require.NotNil(t, couchbaseClient)

		result, err := couchbaseClient.Get()
		require.Error(t, err)
		require.Nil(t, result)
	})
	t.Run("no error", func(t *testing.T) {
		couchbaseClient, err := newCouchbaseClient(componenttest.NewNopHost(), &Config{
			HTTPClientSettings: confighttp.HTTPClientSettings{
				Endpoint: couchbaseNodeStatsMock.URL + "/no_error",
			},
		}, zap.NewNop())
		require.NoError(t, err)
		require.NotNil(t, couchbaseClient)

		result, err := couchbaseClient.Get()
		require.NoError(t, err)
		require.NotNil(t, result)
	})
}

func TestBasicAuth(t *testing.T) {
	username := "otelu"
	password := "otelp"
	encoded := basicAuth(username, password)
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	require.NoError(t, err)
	require.Equal(t, "otelu:otelp", string(decoded))
}
