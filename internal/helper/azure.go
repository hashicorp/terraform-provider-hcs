package helper

import (
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
