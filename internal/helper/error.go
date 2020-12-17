package helper

import "github.com/Azure/go-autorest/autorest/azure"

func IsErrorAzureNotFound(err error) bool {
	azErr, ok := err.(*azure.RequestError)

	return ok && azErr.StatusCode == 404
}
