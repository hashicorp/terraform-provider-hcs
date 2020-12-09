package helper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ParseResourceGroupNameFromID(t *testing.T) {
	tcs := map[string]struct {
		expectErr bool
		expected  string
		input     string
	}{
		"valid id": {
			input:     "/subscriptions/111111/resourceGroups/test-rg-123/foo/bar",
			expected:  "test-rg-123",
			expectErr: false,
		},
		"id too short": {
			input:     "/subscriptions/111111/resourceGroups",
			expectErr: true,
		},
		"malformed id": {
			input:     "/subscriptions/111111/foo/test-rg-132/bar/baz",
			expectErr: true,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			r := require.New(t)

			result, err := ParseResourceGroupNameFromID(tc.input)

			if tc.expectErr {
				r.NotNil(err)
			} else {
				r.NoError(err)
				r.Equal(tc.expected, result)
			}
		})
	}
}
