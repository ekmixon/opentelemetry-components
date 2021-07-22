package rabbitmqreceiver

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

type RabbitMQIntegrationSuite struct {
	suite.Suite
}

func TestRabbitMQIntegration(t *testing.T) {
	suite.Run(t, new(RabbitMQIntegrationSuite))
}

func rabbitmqContainer(t *testing.T) testcontainers.Container {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.rabbitmq",
		},
		ExposedPorts: []string{"15672:15672"},
		WaitingFor:   wait.ForListeningPort("15672"),
	}

	require.NoError(t, req.Validate())

	rabbitmq, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	time.Sleep(time.Second * 6)
	return rabbitmq
}

func (suite *RabbitMQIntegrationSuite) TestRabbitMQScraperHappyPath() {
	t := suite.T()
	rabbitmq := rabbitmqContainer(t)
	defer func() {
		err := rabbitmq.Terminate(context.Background())
		require.NoError(t, err)
	}()
	hostname, err := rabbitmq.Host(context.Background())
	require.NoError(t, err)

	cfg := &Config{
		ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
			CollectionInterval: 100 * time.Millisecond,
		},
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: fmt.Sprintf("http://%s:15672", hostname),
		},
		Password: "dev",
		Username: "dev",
	}

	sc, err := newRabbitMQScraper(zap.NewNop(), cfg)
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

	require.Equal(t, 4, ms.Len())
}
