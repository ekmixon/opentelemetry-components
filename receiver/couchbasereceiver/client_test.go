package couchbasereceiver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.uber.org/zap"
)

func TestGetMetrics(t *testing.T) {
	client, err := newCouchbaseClient(componenttest.NewNopHost(), &Config{
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: "http://127.0.0.1:8091",
		},
		Username: "otelu",
		Password: "otelpassword",
	}, zap.NewNop())
	assert.Nil(t, err)
	stats, err := client.Get()
	assert.Nil(t, err)
	assert.NotNil(t, stats)
}
