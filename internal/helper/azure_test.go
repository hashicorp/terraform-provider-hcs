package helper

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_TagValueToString(t *testing.T) {
	var stringCase interface{} = "test-123"
	var intCase interface{} = 123
	var floatCase interface{} = 1.23

	tcs := map[string]struct {
		expectErr bool
		expected  string
		input     interface{}
	}{
		"valid string": {
			input:     stringCase,
			expected:  "test-123",
			expectErr: false,
		},
		"valid int": {
			input:     intCase,
			expected:  "123",
			expectErr: false,
		},
		"invalid type": {
			input:     floatCase,
			expectErr: true,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			r := require.New(t)

			result, err := TagValueToString(tc.input)
			if tc.expectErr {
				r.Error(err)
			} else {
				r.NoError(err)
				r.Equal(tc.expected, result)
			}
		})
	}
}

// Adapted from the azurerm provider.
// https://github.com/terraform-providers/terraform-provider-azurerm/blob/09bd1119f8628604df0136c66892b28c18d88a06/azurerm/internal/tags/flatten_test.go#L10
func Test_FlattenTags(t *testing.T) {
	testData := []struct {
		Name     string
		Input    map[string]*string
		Expected map[string]interface{}
	}{
		{
			Name:     "Empty",
			Input:    map[string]*string{},
			Expected: map[string]interface{}{},
		},
		{
			Name: "One Item",
			Input: map[string]*string{
				"hello": String("there"),
			},
			Expected: map[string]interface{}{
				"hello": "there",
			},
		},
		{
			Name: "Multiple Items",
			Input: map[string]*string{
				"euros": String("3"),
				"hello": String("there"),
				"panda": String("pops"),
			},
			Expected: map[string]interface{}{
				"euros": "3",
				"hello": "there",
				"panda": "pops",
			},
		},
	}

	for _, v := range testData {
		t.Logf("[DEBUG] Test %q", v.Name)

		actual := FlattenTags(v.Input)
		if !reflect.DeepEqual(actual, v.Expected) {
			t.Fatalf("Expected %+v but got %+v", actual, v.Expected)
		}
	}
}
