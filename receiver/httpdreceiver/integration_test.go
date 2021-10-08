//go:build integration
// +build integration

package httpdreceiver

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/observiq/opentelemetry-components/receiver/helper"
)

func TestHttpdIntegration(t *testing.T) {
	container := getContainer(t, containerRequest2_4)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = fmt.Sprintf("http://%s", net.JoinHostPort(hostname, "8080"))

	expectedFileBytes, err := ioutil.ReadFile("./testdata/examplejsonmetrics/testintegration/expected_metrics.json")
	require.NoError(t, err)

	helper.ValidateIntegrationTestResults(t, cfg, f, expectedFileBytes, map[string]bool{
		"server_name": true,
	})
}

var (
	containerRequest2_4 = testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    path.Join(".", "testdata"),
			Dockerfile: "Dockerfile.httpd",
		},
		ExposedPorts: []string{"8080:80"},
		WaitingFor:   waitStrategy{},
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

type waitStrategy struct{}

func (ws waitStrategy) WaitUntilReady(ctx context.Context, st wait.StrategyTarget) error {
	if err := wait.ForListeningPort("80").
		WithStartupTimeout(2*time.Minute).
		WaitUntilReady(ctx, st); err != nil {
		return err
	}

	hostname, err := st.Host(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			return fmt.Errorf("server startup problem")
		case <-time.After(100 * time.Millisecond):
			resp, err := http.Get(fmt.Sprintf("http://%s:8080/server-status?auto", hostname))
			if err != nil {
				continue
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				continue
			}

			if resp.Body.Close() != nil {
				continue
			}

			// The server needs a moment to generate some stats
			if strings.Contains(string(body), "ReqPerSec") {
				return nil
			}
		}
	}
}
