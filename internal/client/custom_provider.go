package client

import (
	"github.com/Azure/go-autorest/autorest"
)

type CustomProviderClient struct {
	autorest.Client
	BaseURI        string
	SubscriptionID string
}

func NewCustomProviderClientWithBaseURI(baseURI string, subscriptionID string) CustomProviderClient {
	return CustomProviderClient{
		Client:         autorest.NewClientWithUserAgent(""),
		BaseURI:        baseURI,
		SubscriptionID: subscriptionID,
	}
}
