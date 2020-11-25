package client

import (
	"context"
	"fmt"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/go-azure-helpers/authentication"
)

// AzureResourceManagerAccount implementation from the azurerm provider

type AzureResourceManagerAccount struct {
	AuthenticatedAsAServicePrincipal bool
	ClientId                         string
	Environment                      azure.Environment
	ObjectId                         string
	SubscriptionId                   string
	TenantId                         string
}

func NewAzureResourceManagerAccount(ctx context.Context, config authentication.Config, env azure.Environment) (*AzureResourceManagerAccount, error) {
	objectId := ""

	if getAuthenticatedObjectID := config.GetAuthenticatedObjectID; getAuthenticatedObjectID != nil {
		v, err := getAuthenticatedObjectID(ctx)
		if err != nil {
			return nil, fmt.Errorf("error getting authenticated object ID: %v", err)
		}
		objectId = v
	}

	account := AzureResourceManagerAccount{
		AuthenticatedAsAServicePrincipal: config.AuthenticatedAsAServicePrincipal,
		ClientId:                         config.ClientID,
		Environment:                      env,
		ObjectId:                         objectId,
		TenantId:                         config.TenantID,
		SubscriptionId:                   config.SubscriptionID,
	}
	return &account, nil
}
