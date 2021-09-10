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
		require.NotNil(t, err)
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
		require.Nil(t, err)
		require.NotNil(t, couchbaseClient)
	})
}

func TestBasicAuth(t *testing.T) {
	username := "otelu"
	password := "otelp"
	encoded := basicAuth(username, password)
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	require.Nil(t, err)
	require.Equal(t, "otelu:otelp", string(decoded))
}
