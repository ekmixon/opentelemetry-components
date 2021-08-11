package httpdreceiver

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	t.Run("error path", func(t *testing.T) {
		testCases := []struct {
			desc   string
			rawUrl string
			actual error
		}{
			{
				desc:   "only port",
				rawUrl: "8081",
				actual: errors.New("invalid host: expected IP address or localhost, but got 8081"),
			},
			{
				desc:   "invalid host",
				rawUrl: "127.0.0.0.1.1",
				actual: errors.New("invalid host: expected IP address or localhost, but got 127.0.0.0.1.1"),
			},
			{
				desc:   "invalid host, port",
				rawUrl: "127.0.0.0.1.1:8080",
				actual: errors.New("invalid host: expected IP address or localhost, but got 127.0.0.0.1.1"),
			},
			{
				desc:   "invalid path",
				rawUrl: "127.0.0.1:8080/bad-server-status",
				actual: errors.New("invalid path: expected '/server-status' path, but got '/bad-server-status'"),
			},
			{
				desc:   "invalid path, path",
				rawUrl: "127.0.0.1:8080/server-status/server-status",
				actual: errors.New("invalid path: expected '/server-status' path, but got '/server-status/server-status'"),
			},
			{
				desc:   "invalid query",
				rawUrl: "127.0.0.1:8080/server-status?badauto",
				actual: errors.New("invalid raw query: expected 'auto' query, but got 'badauto'"),
			},
			{
				desc:   "invalid frag",
				rawUrl: "127.0.0.1:8080/server-status?auto#badfrag",
				actual: errors.New("invalid fragment: expected empty string, but got badfrag"),
			},
			{
				desc:   "missing colon",
				rawUrl: "127.0.0.18080/server-status?auto",
				actual: errors.New("invalid host: expected IP address or localhost, but got 127.0.0.18080"),
			},
			{
				desc:   "missing colon with space",
				rawUrl: "127.0.0.1 8080/server-status?auto",
				actual: errors.New("invalid endpoint 'http://127.0.0.1 8080/server-status?auto'"),
			},
		}
		for _, tC := range testCases {
			t.Run(tC.desc, func(t *testing.T) {
				cfg := NewFactory().CreateDefaultConfig().(*Config)
				cfg.Endpoint = tC.rawUrl
				require.EqualError(t, cfg.Validate(), tC.actual.Error())
			})
		}
	})

	t.Run("happy path", func(t *testing.T) {
		testCases := []struct {
			desc      string
			rawUrl    string
			actual    error
			parsedUrl string
		}{
			{
				desc:      "default path",
				rawUrl:    "",
				actual:    nil,
				parsedUrl: "http://localhost:8081/server-status?auto",
			},
			{
				desc:      "only host(local)",
				rawUrl:    "localhost",
				actual:    nil,
				parsedUrl: "http://localhost:8081/server-status?auto",
			},
			{
				desc:      "only host",
				rawUrl:    "127.0.0.1",
				actual:    nil,
				parsedUrl: "http://127.0.0.1:8081/server-status?auto",
			},
			{
				desc:      "only host(local), port",
				rawUrl:    "localhost:8081",
				actual:    nil,
				parsedUrl: "http://localhost:8081/server-status?auto",
			},
			{
				desc:      "only host, port",
				rawUrl:    "127.0.0.1:8081",
				actual:    nil,
				parsedUrl: "http://127.0.0.1:8081/server-status?auto",
			},
			{
				desc:      "only host, port, path",
				rawUrl:    "127.0.0.1:8081/server-status",
				actual:    nil,
				parsedUrl: "http://127.0.0.1:8081/server-status?auto",
			},
			{
				desc:      "only host, port, path, query",
				rawUrl:    "127.0.0.1:8081/server-status?auto",
				actual:    nil,
				parsedUrl: "http://127.0.0.1:8081/server-status?auto",
			},
			{
				desc:      "only schema, host, port, path, query",
				rawUrl:    "http://127.0.0.1:8081/server-status?auto",
				actual:    nil,
				parsedUrl: "http://127.0.0.1:8081/server-status?auto",
			},
			{
				desc:      "only schema(https), host, port, path, query",
				rawUrl:    "https://127.0.0.1:8081/server-status?auto",
				actual:    nil,
				parsedUrl: "https://127.0.0.1:8081/server-status?auto",
			},
			{
				desc:      "added slash",
				rawUrl:    "127.0.0.1:8081/server-status/?auto",
				actual:    errors.New("invalid path: expected '/server-status' path, but got '/server-status/'"),
				parsedUrl: "http://127.0.0.1:8081/server-status/?auto",
			},
		}
		for _, tC := range testCases {
			t.Run(tC.desc, func(t *testing.T) {
				cfg := NewFactory().CreateDefaultConfig().(*Config)
				cfg.Endpoint = tC.rawUrl
				require.NoError(t, cfg.Validate())
				require.Equal(t, cfg.Endpoint, tC.parsedUrl)
			})
		}
	})
}
