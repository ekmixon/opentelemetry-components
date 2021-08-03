package elasticsearchreceiver

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	testcases := []struct {
		name        string
		expectErr   bool
		config      Config
		expectedErr string
	}{
		{
			name:      "valid config",
			expectErr: false,
			config: func() Config {
				config := Config{}
				config.Password = "dev"
				config.Username = "devname"
				config.Endpoint = "http://localhost:9200"
				return config
			}(),
		},
		{
			name:      "missing password",
			expectErr: true,
			config: func() Config {
				config := Config{}
				config.Username = "devname"
				config.Endpoint = "http://localhost:9200"
				return config
			}(),
			expectedErr: "'username' specified but not 'password'",
		},
		{
			name:      "missing username",
			expectErr: true,
			config: func() Config {
				config := Config{}
				config.Password = "dev"
				config.Endpoint = "http://localhost:9200"
				return config
			}(),
			expectedErr: "'password' specified but not 'username'",
		},
		{
			name:      "bad endpoint",
			expectErr: true,
			config: func() Config {
				config := Config{}
				config.Endpoint = "https://this shouldn't work"
				return config
			}(),
			expectedErr: "invalid url specified in field 'endpoint'",
		},
		{
			name:      "missing endpoint",
			expectErr: true,
			config: func() Config {
				return Config{}
			}(),
			expectedErr: "missing required field 'endpoint'",
		},
	}
	for _, tc := range testcases {
		t.Run("elasticsearch/configtest/"+tc.name, func(t *testing.T) {
			if tc.expectErr {
				err := tc.config.Validate()
				require.NotNil(t, err)
				require.Equal(t, tc.expectedErr, err.Error())
			} else {
				require.NoError(t, tc.config.Validate())
			}
		})
	}
}
