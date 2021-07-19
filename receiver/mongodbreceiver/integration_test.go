// +build integration

package mongodbreceiver

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/model/pdata"
)

func TestIntegration(t *testing.T) {
	// cs := container.New(t)
	// c := cs.StartImage("mongodb:1.6-alpine", container.WithPortReady(27017))

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = net.JoinHostPort("localhost", "27017")

	user := "otel"
	pass := "otel"
	cfg.User = &user
	cfg.Password = &pass

	consumer := new(consumertest.MetricsSink)

	rcvr, err := f.CreateMetricsReceiver(context.Background(), componenttest.NewNopReceiverCreateSettings(), cfg, consumer)
	require.NoError(t, err, "failed creating metrics receiver")
	require.NoError(t, rcvr.Start(context.Background(), componenttest.NewNopHost()))

	require.Eventuallyf(t, func() bool {
		return len(consumer.AllMetrics()) > 0
	}, 15*time.Second, 1*time.Second, "failed to receive more than 0 metrics")

	md := consumer.AllMetrics()[0]

	require.Equal(t, 1, md.ResourceMetrics().Len())

	ilms := md.ResourceMetrics().At(0).InstrumentationLibraryMetrics()
	require.Equal(t, 1, ilms.Len())

	metrics := ilms.At(0).Metrics()
	require.Equal(t, 13, metrics.Len())

	assertAllMetricNamesArePresent(t, metadata.Metrics.Names(), metrics)

	assert.NoError(t, rcvr.Shutdown(context.Background()))
}

func assertAllMetricNamesArePresent(t *testing.T, names []string, metrics pdata.MetricSlice) {
	seen := make(map[string]bool, len(names))
	for i := range names {
		seen[names[i]] = false
	}

	for i := 0; i < metrics.Len(); i++ {
		seen[metrics.At(i).Name()] = true
	}

	for k, v := range seen {
		if !v {
			t.Fatalf("Did not find metric %q", k)
		}
	}
}
