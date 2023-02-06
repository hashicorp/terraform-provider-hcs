// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package hcsmeta

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_RegionIsSupported(t *testing.T) {
	supportedRegions := []SupportedRegion{
		{
			ShortName:    "uswest2",
			FriendlyName: "West US",
		},
		{
			ShortName:    "westeurope",
			FriendlyName: "West Europe",
		},
	}

	tcs := map[string]struct {
		expected         bool
		input            string
		supportedRegions []SupportedRegion
	}{
		"supported region": {
			input:            "uswest2",
			expected:         true,
			supportedRegions: supportedRegions,
		},
		"unsupported region": {
			input:            "francecentral",
			expected:         false,
			supportedRegions: supportedRegions,
		},
		"nil supported regions": {
			input:            "francecentral",
			expected:         true,
			supportedRegions: nil,
		},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			r := require.New(t)

			result := RegionIsSupported(tc.input, tc.supportedRegions)
			r.Equal(tc.expected, result)
		})
	}
}
