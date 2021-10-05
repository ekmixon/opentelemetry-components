//go:build integration
// +build integration

package couchdbreceiver

import (
	"context"
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

func TestCouchdbIntegration(t *testing.T) {
	t.Run("Running docker version 3.0.0 on port 5985", func(t *testing.T) {
		container := getContainer(t, containerRequest300)
		defer func() {
			require.NoError(t, container.Terminate(context.Background()))
		}()
		hostname, err := container.Host(context.Background())
		require.NoError(t, err)

		f := NewFactory()
		cfg := f.CreateDefaultConfig().(*Config)
		cfg.Endpoint = net.JoinHostPort(hostname, "5984")
		cfg.Username = "otel"
		cfg.Password = "otel"
		cfg.Nodename = "_local"

		expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testintegration/expected_metrics.json")
		require.NoError(t, err)

		helper.IntegrationTestHelper(t, cfg, f, expectedFileBytes, map[string]bool{})
	})

	t.Run("Running docker version 3.1.1 on port 5984", func(t *testing.T) {
		container := getContainer(t, containerRequest311)
		defer func() {
			require.NoError(t, container.Terminate(context.Background()))
		}()
		hostname, err := container.Host(context.Background())
		require.NoError(t, err)

		f := NewFactory()
		cfg := f.CreateDefaultConfig().(*Config)
		cfg.Endpoint = net.JoinHostPort(hostname, "5984")
		cfg.Username = "otel"
		cfg.Password = "otel"
		cfg.Nodename = "_local"

		expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testintegration/expected_metrics.json")
		require.NoError(t, err)

		helper.IntegrationTestHelper(t, cfg, f, expectedFileBytes, map[string]bool{})
	})
}

var (
	containerRequest300 = testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.couchdb.300",
		},
		ExposedPorts: []string{"5984:5984"}, // TODO: standardize how to do ports between local-targets
		WaitingFor: wait.ForListeningPort("5984").
			WithStartupTimeout(2 * time.Minute),
	}
	containerRequest311 = testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.couchdb.311",
		},
		ExposedPorts: []string{"5984:5984"},
		WaitingFor: wait.ForListeningPort("5984").
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
