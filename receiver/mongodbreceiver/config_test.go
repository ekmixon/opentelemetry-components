package mongodbreceiver

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config/configtls"
)

func TestValidate(t *testing.T) {
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
			expected: errors.New("password provided without user"),
		},
		{
			desc:     "with username, no password",
			username: "user",
			password: "",
			expected: errors.New("user provided without password"),
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
			cfg := Config{Username: tC.username, Password: tC.password}
			err := cfg.Validate()
			if tC.expected == nil {
				require.Nil(t, err)
			} else {
				require.EqualError(t, tC.expected, err.Error())
			}
		})
	}
}

func TestBadTLSConfigs(t *testing.T) {
	testCases := []struct {
		tlsConfig   configtls.TLSClientSetting
		expectError bool
	}{
		{
			tlsConfig: configtls.TLSClientSetting{
				TLSSetting: configtls.TLSSetting{
					CAFile: "not/a/real/file.pem",
				},
				Insecure:           false,
				InsecureSkipVerify: false,
				ServerName:         "",
			},
			expectError: true,
		},
		{
			tlsConfig: configtls.TLSClientSetting{
				TLSSetting:         configtls.TLSSetting{},
				Insecure:           false,
				InsecureSkipVerify: false,
				ServerName:         "",
			},
			expectError: false,
		},
	}
	for _, tc := range testCases {
		cfg := Config{Username: "otel", Password: "pword", TLSClientSetting: tc.tlsConfig}
		err := cfg.Validate()
		if tc.expectError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}
