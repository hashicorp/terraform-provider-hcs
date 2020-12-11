package helper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-provider-hcs/internal/clients/hcs-ama-api-spec/models"
)

func Test_AMAVersionStatusToString(t *testing.T) {
	tcs := map[string]struct {
		expected string
		input    models.HashicorpCloudConsulamaAmaVersionStatus
	}{
		"available": {
			input:    models.HashicorpCloudConsulamaAmaVersionStatusAVAILABLE,
			expected: "AVAILABLE",
		},
		"recommended ": {
			input:    models.HashicorpCloudConsulamaAmaVersionStatusRECOMMENDED,
			expected: "RECOMMENDED",
		},
		"preview": {
			input:    models.HashicorpCloudConsulamaAmaVersionStatusPREVIEW,
			expected: "PREVIEW",
		},
		"default": {
			input:    "FOO",
			expected: "",
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			r := require.New(t)

			result := AMAVersionStatusToString(tc.input)

			r.Equal(tc.expected, result)
		})
	}
}
