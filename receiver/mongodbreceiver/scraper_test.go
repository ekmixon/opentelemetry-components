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
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.uber.org/zap"
)

func TestScraper(t *testing.T) {
	//cs := container.New(t)
	//c := cs.StartImage("mongo:4.0.25-xenial", container.WithPortReady(27017))

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = net.JoinHostPort("localhost", "27017")

	user := "otel"
	pass := "otel"
	cfg.User = &user
	cfg.Password = &pass

	sc := newMongodbScraper(zap.NewNop(), cfg)

	err := sc.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)
	metrics, err := sc.Scrape(context.Background(), cfg.ID())
	require.Nil(t, err)
	rms := metrics.ResourceMetrics()
	require.Equal(t, 1, rms.Len())
	rm := rms.At(0)

	ilms := rm.InstrumentationLibraryMetrics()
	require.Equal(t, 1, ilms.Len())

	ilm := ilms.At(0)
	ms := ilm.Metrics()

	require.Equal(t, 7, ms.Len())

	for i := 0; i < ms.Len(); i++ {
		m := ms.At(i)
		switch m.Name() {
		case "mongodb.collections":
			require.Equal(t, 3, m.IntGauge().DataPoints().Len())
		case "mongodb.data_size":
			require.Equal(t, 3, m.Gauge().DataPoints().Len())
		case "mongodb.extents":
			require.Equal(t, 3, m.IntGauge().DataPoints().Len())
		case "mongodb.index_size":
			require.Equal(t, 3, m.Gauge().DataPoints().Len())
		case "mongodb.indexes":
			require.Equal(t, 3, m.IntGauge().DataPoints().Len())
		case "mongodb.objects":
			require.Equal(t, 3, m.IntGauge().DataPoints().Len())
		case "mongodb.storage_size":
			require.Equal(t, 3, m.Gauge().DataPoints().Len())
		default:
			t.Error("Incorrect name or untracked metric name.")
		}
	}
}
