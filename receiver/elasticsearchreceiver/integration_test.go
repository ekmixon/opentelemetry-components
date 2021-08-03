// +build integration

package elasticsearchreceiver

import (
	"context"
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/receiver/scraperhelper"
	"go.uber.org/zap"
)

type ElasticSearchIntegrationSuite struct {
	suite.Suite
}

func TestElasticSearchIntegration(t *testing.T) {
	suite.Run(t, new(ElasticSearchIntegrationSuite))
}

func elasticsearchContainer_7_8(t *testing.T) testcontainers.Container {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.elasticsearch_7.8",
		},
		ExposedPorts: []string{"9200:9200"},
		WaitingFor:   wait.ForListeningPort("9200"),
	}

	require.NoError(t, req.Validate())

	elasticsearch, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	time.Sleep(time.Second * 6)
	return elasticsearch
}

func elasticsearchContainer_7_13(t *testing.T) testcontainers.Container {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.elasticsearch_7.13",
		},
		ExposedPorts: []string{"9200:9200"},
		WaitingFor:   wait.ForListeningPort("9200"),
	}

	require.NoError(t, req.Validate())

	elasticsearch, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	time.Sleep(time.Second * 6)
	return elasticsearch
}

func (suite *ElasticSearchIntegrationSuite) TestElasticSearchScraperHappyPath7_8() {
	t := suite.T()
	elasticsearch := elasticsearchContainer_7_8(t)
	defer func() {
		err := elasticsearch.Terminate(context.Background())
		require.NoError(t, err)
	}()
	hostname, err := elasticsearch.Host(context.Background())
	require.NoError(t, err)

	cfg := &Config{
		ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
			CollectionInterval: 100 * time.Millisecond,
		},
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: fmt.Sprintf("http://%s:9200", hostname),
		},
	}

	sc, err := newElasticSearchScraper(zap.NewNop(), cfg)
	require.NoError(t, err)
	err = sc.start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)
	rms, err := sc.scrape(context.Background())

	require.NoError(t, err)

	require.Equal(t, 1, rms.Len())
	rm := rms.At(0)

	ilms := rm.InstrumentationLibraryMetrics()
	require.Equal(t, 1, ilms.Len())

	ilm := ilms.At(0)
	ms := ilm.Metrics()

	require.Equal(t, 17, ms.Len())
}

func (suite *ElasticSearchIntegrationSuite) TestElasticSearchScraperHappyPath7_13() {
	t := suite.T()
	elasticsearch := elasticsearchContainer_7_13(t)
	defer func() {
		err := elasticsearch.Terminate(context.Background())
		require.NoError(t, err)
	}()
	hostname, err := elasticsearch.Host(context.Background())
	require.NoError(t, err)

	cfg := &Config{
		ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
			CollectionInterval: 100 * time.Millisecond,
		},
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: fmt.Sprintf("http://%s:9200", hostname),
		},
	}

	sc, err := newElasticSearchScraper(zap.NewNop(), cfg)
	require.NoError(t, err)
	err = sc.start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)
	rms, err := sc.scrape(context.Background())

	require.NoError(t, err)

	require.Equal(t, 1, rms.Len())
	rm := rms.At(0)

	ilms := rm.InstrumentationLibraryMetrics()
	require.Equal(t, 1, ilms.Len())

	ilm := ilms.At(0)
	ms := ilm.Metrics()

	require.Equal(t, 17, ms.Len())
}
