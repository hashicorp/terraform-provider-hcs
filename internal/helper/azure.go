package helper

import (
	"fmt"

	"github.com/Azure/go-autorest/autorest"
)

// IsAutoRestResponseCodeNotFound determines if an AutoRest response code was
// 404 not found.
func IsAutoRestResponseCodeNotFound(resp autorest.Response) bool {
	if r := resp.Response; r != nil {
		if r.StatusCode == 404 {
			return true
		}
	}

	return false
}

// TagValueToString converts a tag interface{} to string.
// Adapted from the azurerm provider.
// https://github.com/terraform-providers/terraform-provider-azurerm/blob/b7299d0b8c6f3685db07586530a7f52216dd48e4/azurerm/internal/tags/validation.go#L31
func TagValueToString(v interface{}) (string, error) {
	switch value := v.(type) {
	case string:
		return value, nil
	case int:
		return fmt.Sprintf("%d", value), nil
	default:
		return "", fmt.Errorf("unknown tag type %T in tag value", value)
	}
}

// FlattenTags converts a tag map of *string values to interface{} values.
// Adapted from the azurerm provider.
// https://github.com/terraform-providers/terraform-provider-azurerm/blob/7a46303711d53414249b1829d6d879a5dbdae9c4/azurerm/internal/tags/flatten.go#L9
func FlattenTags(tagMap map[string]*string) map[string]interface{} {
	// If tagsMap is nil, len(tagsMap) will be 0.
	output := make(map[string]interface{}, len(tagMap))

	for i, v := range tagMap {
		if v == nil {
			continue
		}

		output[i] = *v
	}

	return output
}
