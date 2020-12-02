package client

import (
	"log"
	"sync"

	"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/go-uuid"
)

// Correlation request id implementation from the azurerm provider
// https://github.com/terraform-providers/terraform-provider-azurerm/blob/8f32ad645888ee00a24ad7c739a8703222e13913/azurerm/internal/common/correlation_id.go

const (
	// HeaderCorrelationRequestID is the Azure extension header to set a user-specified correlation request ID.
	HeaderCorrelationRequestID = "x-ms-correlation-request-id"
)

var (
	// msCorrelationRequestIDOnce is used to maintain the same correlation id throughout all Azure requests.
	msCorrelationRequestIDOnce sync.Once
	// msCorrelationRequestID the correlation id to use throughout all Azure requests.
	msCorrelationRequestID string
)

// withCorrelationRequestID returns a PrepareDecorator that adds an HTTP extension header of
// `x-ms-correlation-request-id` whose value is passed, undecorated UUID (e.g.,7F5A6223-F475-4A9C-B9D5-12575AA6B11B`).
func withCorrelationRequestID(uuid string) autorest.PrepareDecorator {
	return autorest.WithHeader(HeaderCorrelationRequestID, uuid)
}

// correlationRequestID generates an UUID to pass through `x-ms-correlation-request-id` header.
func correlationRequestID() string {
	msCorrelationRequestIDOnce.Do(func() {
		var err error
		msCorrelationRequestID, err = uuid.GenerateUUID()

		if err != nil {
			log.Printf("[WARN] Failed to generate uuid for msCorrelationRequestID: %+v", err)
		}

		log.Printf("[DEBUG] Genereated Provider Correlation Request Id: %s", msCorrelationRequestID)
	})

	return msCorrelationRequestID
}
