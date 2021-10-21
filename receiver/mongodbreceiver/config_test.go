package mongodbreceiver

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
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
