package helper

import (
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

func IsErrorAzureNotFound(err error) bool {
	azErr, ok := err.(*azure.RequestError)

	return ok && azErr.StatusCode == 404
}

func IsResponseCodeNotFound(resp autorest.Response) bool {
	if r := resp.Response; r != nil {
		if r.StatusCode == 404 {
			return true
		}
	}

	return false
}
