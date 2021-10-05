//go:build integration
// +build integration

package elasticsearchreceiver

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

func TestElasticSearch7_8(t *testing.T) {
	container := getContainer(t, containerRequest7_8)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = fmt.Sprintf("http://%s", net.JoinHostPort(hostname, "9200"))

	expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testintegration7_8/expected_metrics.json")
	require.NoError(t, err)

	helper.IntegrationTestHelper(t, cfg, f, expectedFileBytes, map[string]bool{
		"server_name": true,
	})
}

func TestElasticSearch7_13(t *testing.T) {
	container := getContainer(t, containerRequest7_13)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = fmt.Sprintf("http://%s", net.JoinHostPort(hostname, "9200"))

	expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testintegration7_13/expected_metrics.json")
	require.NoError(t, err)

	helper.IntegrationTestHelper(t, cfg, f, expectedFileBytes, map[string]bool{
		"server_name": true,
	})
}

var (
	containerRequest7_8 = testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.elasticsearch_7.8",
		},
		ExposedPorts: []string{"9200:9200"},
		WaitingFor: wait.ForListeningPort("9200").
			WithStartupTimeout(2 * time.Minute),
	}
	containerRequest7_13 = testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.elasticsearch_7.13",
		},
		ExposedPorts: []string{"9200:9200"},
		WaitingFor: wait.ForListeningPort("9200").
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
	time.Sleep(time.Second * 6) // TODO custom wait.Strategy
	return container
}
