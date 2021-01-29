package clients

import (
	"context"
	"fmt"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/go-azure-helpers/authentication"
)

// ResourceManagerAccount implementation from the azurerm provider
// https://github.com/terraform-providers/terraform-provider-azurerm/blob/8f32ad645888ee00a24ad7c739a8703222e13913/azurerm/internal/clients/auth.go

// AzureResourceManagerAccount contains Azure AD account information.
type AzureResourceManagerAccount struct {
	// AuthenticatedAsAServicePrincipal denotes whether the current Azure AD user is authenticated using a SP.
	AuthenticatedAsAServicePrincipal bool
	// ClientId is the client id of the Azure AD account.
	ClientId string
	// Environment represents a set of endpoints for each of Azure's Clouds.
	Environment azure.Environment
	// ObjectId is the id of the Azure AD account.
	ObjectId string
	// SubscriptionId is the id of the Azure subscription.
	SubscriptionId string
	// TenantId is the id of the Azure tenant.
	TenantId string
}

// NewAzureResourceManagerAccount constructs an AzureResourceManagerAccount from the given auth config and Azure env.
func NewAzureResourceManagerAccount(ctx context.Context, config authentication.Config, env azure.Environment) (*AzureResourceManagerAccount, error) {
	objectId := ""

	if getAuthenticatedObjectID := config.GetAuthenticatedObjectID; getAuthenticatedObjectID != nil {
		v, err := getAuthenticatedObjectID(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch authenticated object ID: %v", err)
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
