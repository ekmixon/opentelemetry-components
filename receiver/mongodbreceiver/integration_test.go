//go:build integration
// +build integration

package mongodbreceiver

import (
	"context"
	"io/ioutil"
	"net"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/observiq/opentelemetry-components/receiver/helper"
)

func TestMongoDBIntegration(t *testing.T) {
	container := getContainer(t, containerRequest4_0)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = net.JoinHostPort(hostname, "37017")
	cfg.Username = "otel"
	cfg.Password = "otel"

	expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testintegration/expected_metrics.json")
	require.NoError(t, err)

	helper.ValidateIntegrationTestResults(t, cfg, f, expectedFileBytes, map[string]bool{})
}

var (
	containerRequest4_0 = testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.mongodb",
		},
		ExposedPorts: []string{"37017:27017"},
		WaitingFor: wait.ForListeningPort("27017").
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
	return container
}
