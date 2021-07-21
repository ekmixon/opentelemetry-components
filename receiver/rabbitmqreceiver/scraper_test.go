package rabbitmqreceiver

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/observiq/opentelemetry-components/receiver/rabbitmqreceiver/internal/metadata"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configtls"
	"go.uber.org/zap"
)

func TestScraper(t *testing.T) {
	rabbitmqMock := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/api/queues" {
			rw.WriteHeader(200)
			_, _ = rw.Write([]byte(`
[{"arguments":{},"auto_delete":false,"backing_queue_status":{"avg_ack_egress_rate":1.0113543644652656,"avg_ack_ingress_rate":1.0113543644652656,"avg_egress_rate":1.0113543644652656,"avg_ingress_rate":0.9878865274095732,"delta":["delta",6144,8199,0,14343],"len":11497,"mode":"default","next_seq_id":17298,"q1":25,"q2":2930,"q3":342,"q4":1,"target_ram_count":4146},"consumer_capacity":1.1197605996721166e-4,"consumer_utilisation":1.1197541246445536e-4,"consumers":1,"durable":true,"effective_policy_definition":{},"exclusive":false,"exclusive_consumer_tag":null,"garbage_collection":{"fullsweep_after":65535,"max_heap_size":0,"min_bin_vheap_size":46422,"min_heap_size":233,"minor_gcs":0},"head_message_timestamp":null,"memory":4120816,"message_bytes":2311098,"message_bytes_paged_out":0,"message_bytes_persistent":2311098,"message_bytes_ram":663099,"message_bytes_ready":2310897,"message_bytes_unacknowledged":201,"message_stats":{"ack":5800,"ack_details":{"rate":2.0},"deliver":5803,"deliver_details":{"rate":2.0},"deliver_get":5803,"deliver_get_details":{"rate":2.0},"deliver_no_ack":0,"deliver_no_ack_details":{"rate":0.0},"get":0,"get_details":{"rate":0.0},"get_empty":0,"get_empty_details":{"rate":0.0},"get_no_ack":0,"get_no_ack_details":{"rate":0.0},"publish":17299,"publish_details":{"rate":2.0},"redeliver":2,"redeliver_details":{"rate":0.0}},"messages":11503,"messages_details":{"rate":0.0},"messages_paged_out":0,"messages_persistent":11498,"messages_ram":3299,"messages_ready":11502,"messages_ready_details":{"rate":0.0},"messages_ready_ram":3298,"messages_unacknowledged":1,"messages_unacknowledged_details":{"rate":0.0},"messages_unacknowledged_ram":1,"name":"webq1","node":"rabbit@43dcde7ee8aa","operator_policy":null,"policy":null,"recoverable_slaves":null,"reductions":23085306,"reductions_details":{"rate":3318.2},"single_active_consumer_tag":null,"state":"running","type":"classic","vhost":"dev"}]
			`))
			return
		}
		rw.WriteHeader(404)
	}))
	sc := newRabbitMQScraper(zap.NewNop(), &Config{
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: rabbitmqMock.URL,
		},
	})
	err := sc.start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)
	rms, err := sc.scrape(context.Background())
	require.Nil(t, err)

	require.Equal(t, 1, rms.Len())
	rm := rms.At(0)

	ilms := rm.InstrumentationLibraryMetrics()
	require.Equal(t, 1, ilms.Len())

	ilm := ilms.At(0)
	ms := ilm.Metrics()

	require.Equal(t, 4, ms.Len())

	metricValues := make(map[string]float64, 7)

	for i := 0; i < ms.Len(); i++ {
		m := ms.At(i)

		dps := m.Gauge().DataPoints()
		if dps.Len() > 1 {
			for j := 0; j < dps.Len(); j++ {
				dp := dps.At(j)
				state, _ := dp.LabelsMap().Get(metadata.L.State)
				label := fmt.Sprintf("%s state:%s", m.Name(), state)
				metricValues[label] = dp.Value()
			}
		} else {
			dp := dps.At(0)
			metricValues[m.Name()] = dp.Value()
		}
	}

	require.Equal(t, map[string]float64{
		"rabbitmq.publish_rate":                      2,
		"rabbitmq.delivery_rate":                     2,
		"rabbitmq.consumers":                         1,
		"rabbitmq.num_messages state:ready":          11502,
		"rabbitmq.num_messages state:total":          11503,
		"rabbitmq.num_messages state:unacknowledged": 1,
	}, metricValues)
}

func TestScraperFailedStart(t *testing.T) {
	sc := newRabbitMQScraper(zap.NewNop(), &Config{
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: "localhost:8080",
			TLSSetting: configtls.TLSClientSetting{
				TLSSetting: configtls.TLSSetting{
					CAFile: "/non/existent",
				},
			},
		},
	})
	err := sc.start(context.Background(), componenttest.NewNopHost())
	require.Error(t, err)
}
