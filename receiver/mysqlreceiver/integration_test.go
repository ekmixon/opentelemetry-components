//go:build integration
// +build integration

package mysqlreceiver

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.uber.org/zap"

	"github.com/observiq/opentelemetry-components/receiver/helper"
)

func TestMysqlIntegration(t *testing.T) {
	t.Run("Running mysql version 5.7", func(t *testing.T) {
		container := getContainer(t, containerRequest5_7)
		defer func() {
			require.NoError(t, container.Terminate(context.Background()))
		}()
		hostname, err := container.Host(context.Background())
		require.NoError(t, err)

		f := NewFactory()
		cfg := f.CreateDefaultConfig().(*Config)
		cfg.Endpoint = net.JoinHostPort(hostname, "3307")
		cfg.Username = "otel"
		cfg.Password = "otel"

		expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testintegration/expected_metrics.json")
		require.NoError(t, err)

		helper.IntegrationTestHelper(t, cfg, f, expectedFileBytes, map[string]bool{})
	})

	t.Run("Running mysql version 8.0", func(t *testing.T) {
		container := getContainer(t, containerRequest8_0)
		defer func() {
			require.NoError(t, container.Terminate(context.Background()))
		}()
		hostname, err := container.Host(context.Background())
		require.NoError(t, err)

		f := NewFactory()
		cfg := f.CreateDefaultConfig().(*Config)
		cfg.Endpoint = net.JoinHostPort(hostname, "3306")
		cfg.Username = "otel"
		cfg.Password = "otel"

		expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testintegration/expected_metrics.json")
		require.NoError(t, err)

		helper.IntegrationTestHelper(t, cfg, f, expectedFileBytes, map[string]bool{})
	})
}

var (
	containerRequest5_7 = testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.mysql.5_7",
		},
		ExposedPorts: []string{"3307:3306"},
		WaitingFor: wait.ForListeningPort("3306").
			WithStartupTimeout(2 * time.Minute),
	}
	containerRequest8_0 = testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.mysql.8_0",
		},
		ExposedPorts: []string{"3306:3306"},
		WaitingFor: wait.ForListeningPort("3306").
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
	return container
}

func TestMySQLStartStop(t *testing.T) {
	container := getContainer(t, containerRequest8_0)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	sc := newMySQLScraper(zap.NewNop(), &Config{
		Username: "otel",
		Password: "otel",
		Endpoint: fmt.Sprintf("%s:3306", hostname),
	})

	// scraper is connected
	err = sc.start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	// scraper is closed
	err = sc.shutdown(context.Background())
	require.NoError(t, err)

	// scraper scapes and produces an error because db is closed
	_, err = sc.scrape(context.Background())
	require.Error(t, err)
	require.EqualError(t, err, "sql: database is closed")
}
