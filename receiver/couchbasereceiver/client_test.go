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
			Endpoint: "http://127.0.0.1:8091/pools/default/stats/range",
		},
		Username: "Administrator",
		Password: "password",
	}, zap.NewNop())
	assert.Nil(t, err)
	m := Metrics
	newMetrics := client.Post(m)
	assert.Nil(t, newMetrics)
}
