//go:build integration
// +build integration

package rabbitmqreceiver

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"path"
	"testing"
	"time"

	"github.com/observiq/opentelemetry-components/receiver/helper"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRabbitMQScraperHappyPath(t *testing.T) {
	container := getContainer(t, containerRequest3_8)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = fmt.Sprintf("http://%s", net.JoinHostPort(hostname, "15672"))
	cfg.Password = "dev"
	cfg.Username = "dev"

	expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testintegration/expected_metrics.json")
	require.NoError(t, err)

	helper.IntegrationTestHelper(t, cfg, f, expectedFileBytes, map[string]bool{})
}

var (
	containerRequest3_8 = testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.rabbitmq",
		},
		ExposedPorts: []string{"15672:15672"},
		WaitingFor: wait.ForListeningPort("15672").
			WithStartupTimeout(2 * time.Minute),
	}
)

func getContainer(t *testing.T, req testcontainers.ContainerRequest) testcontainers.Container {
	require.NoError(t, req.Validate())
	container, err := testcontainers.GenericContainer(
		context.Background(),
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
	require.NoError(t, err)

	code, err := container.Exec(context.Background(), []string{"/setup.sh"})
	require.NoError(t, err)
	require.Equal(t, 0, code)
	time.Sleep(time.Second * 6) // TODO customize wait.Strategy
	return container
}
