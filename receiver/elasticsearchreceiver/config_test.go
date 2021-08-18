package elasticsearchreceiver

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	t.Run("error path, bad url", func(t *testing.T) {
		cfg := NewFactory().CreateDefaultConfig().(*Config)
		cfg.Endpoint = "http://localhost:9200"
		cfg.Username = "user"
		require.Equal(t, errors.New("'username' specified but not 'password'"), cfg.Validate())
	})
	t.Run("error path, bad credentials", func(t *testing.T) {
		cfg := NewFactory().CreateDefaultConfig().(*Config)
		cfg.Endpoint = "http://endpoint with space"
		require.Equal(t, errors.New("invalid endpoint 'http://endpoint with space'"), cfg.Validate())
	})

	t.Run("happy path", func(t *testing.T) {
		testCases := []struct {
			desc        string
			rawUrl      string
			expected    error
			expectedUrl string
		}{
			{
				desc:        "default path",
				rawUrl:      "",
				expected:    nil,
				expectedUrl: "http://localhost:9200",
			},
			{
				desc:        "only host(local)",
				rawUrl:      "localhost",
				expected:    nil,
				expectedUrl: "http://localhost:9200",
			},
			{
				desc:        "only host",
				rawUrl:      "127.0.0.1",
				expected:    nil,
				expectedUrl: "http://127.0.0.1:9200",
			},
			{
				desc:        "host(local) and port",
				rawUrl:      "localhost:9200",
				expected:    nil,
				expectedUrl: "http://localhost:9200",
			},
			{
				desc:        "host and port",
				rawUrl:      "127.0.0.1:9200",
				expected:    nil,
				expectedUrl: "http://127.0.0.1:9200",
			},
			{
				desc:        "unique host no port",
				rawUrl:      "myAlias.Site",
				expected:    nil,
				expectedUrl: "http://myAlias.Site:9200",
			},
			{
				desc:        "unique host with port",
				rawUrl:      "myAlias.Site:1234",
				expected:    nil,
				expectedUrl: "http://myAlias.Site:1234",
			},
			{
				desc:        "only port",
				rawUrl:      ":9200",
				expected:    nil,
				expectedUrl: "http://localhost:9200",
			},
			{
				desc:        "limitation: double port",
				rawUrl:      "1234",
				expected:    nil,
				expectedUrl: "http://1234:9200",
			},
			{
				desc:        "limitation: invalid ip",
				rawUrl:      "500.0.0.0.1.1",
				expected:    nil,
				expectedUrl: "http://500.0.0.0.1.1:9200",
			},
		}
		for _, tC := range testCases {
			t.Run(tC.desc, func(t *testing.T) {
				cfg := NewFactory().CreateDefaultConfig().(*Config)
				cfg.Endpoint = tC.rawUrl
				require.NoError(t, cfg.Validate())
				require.Equal(t, tC.expectedUrl, cfg.Endpoint)
			})
		}
	})
}

func TestMissingProtocol(t *testing.T) {
	testCases := []struct {
		desc     string
		proto    string
		expected bool
	}{
		{
			desc:     "http proto",
			proto:    "http://localhost",
			expected: false,
		},
		{
			desc:     "https proto",
			proto:    "https://localhost",
			expected: false,
		},
		{
			desc:     "HTTP caps",
			proto:    "HTTP://localhost",
			expected: false,
		},
		{
			desc:     "everything else",
			proto:    "ht",
			expected: true,
		},
		{
			desc:     "everything else",
			proto:    "localhost",
			expected: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			require.Equal(t, tC.expected, missingProtocol(tC.proto))
		})
	}
}

func TestInvalidCredentials(t *testing.T) {
	testCases := []struct {
		desc     string
		username string
		password string
		expected error
	}{
		{
			desc:     "no username, no password",
			username: "",
			password: "",
			expected: nil,
		},
		{
			desc:     "no username, with password",
			username: "",
			password: "pass",
			expected: errors.New("'password' specified but not 'username'"),
		},
		{
			desc:     "with username, no password",
			username: "user",
			password: "",
			expected: errors.New("'username' specified but not 'password'"),
		},
		{
			desc:     "no username or password",
			username: "user",
			password: "pass",
			expected: nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := invalidCredentials(tC.username, tC.password)
			if tC.expected == nil {
				require.Nil(t, err)
			} else {
				require.EqualError(t, tC.expected, err.Error())
			}
		})
	}
}
