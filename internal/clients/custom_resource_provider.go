package clients

import (
	"context"
	"net/http"

	"github.com/Azure/go-autorest/autorest/azure"

	"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/terraform-provider-hcs/internal/clients/hcs-ama-api-spec/models"
)

// CustomResourceProviderClient is used to make authenticated requests to the HCS Azure Custom Resource Provider.
type CustomResourceProviderClient struct {
	// Client is the Autorest client responsible for making HTTP requests to Azure.
	autorest.Client
	// BaseURI is the base URI for the Azure Management API.
	BaseURI string
	// SubscriptionID is the Azure subscription id for the current authenticated user.
	SubscriptionID string
}

// NewCustomResourceProviderClientWithBaseURI constructs a CustomResourceProviderClient using the provided
// base URI and subscription id.
func NewCustomResourceProviderClientWithBaseURI(baseURI string, subscriptionID string) CustomResourceProviderClient {
	return CustomResourceProviderClient{
		Client:         autorest.NewClientWithUserAgent("hcs-custom-resource-provider"),
		BaseURI:        baseURI,
		SubscriptionID: subscriptionID,
	}
}

// customActionPreparer prepares the Custom Resource Action request.
func (client CustomResourceProviderClient) customActionPreparer(ctx context.Context, managedResourceGroupId string, action string, body interface{}) (*http.Request, error) {
	pathParams := map[string]interface{}{
		"resourceGroup":  autorest.Encode("path", managedResourceGroupId),
		"action":         autorest.Encode("path", action),
		"subscriptionId": autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2018-09-01-preview"
	queryParams := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsPost(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroup}/providers/Microsoft.CustomProviders/resourceProviders/public/{action}", pathParams),
		autorest.WithJSON(body),
		autorest.WithQueryParameters(queryParams))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// CreateRootToken invokes the createToken Custom Resource Action.
func (client CustomResourceProviderClient) CreateRootToken(ctx context.Context, managedResourceGroupId string) (models.HashicorpCloudConsulamaAmaCreateTokenResponse, error) {
	var rootToken models.HashicorpCloudConsulamaAmaCreateTokenResponse

	req, err := client.customActionPreparer(ctx, managedResourceGroupId, "createToken", models.HashicorpCloudConsulamaAmaCreateTokenRequest{
		ResourceGroup:  managedResourceGroupId,
		SubscriptionID: client.SubscriptionID,
	})
	if err != nil {
		return rootToken, err
	}

	var resp *http.Response
	resp, err = client.Send(req, azure.DoRetryWithRegistration(client.Client))
	if err != nil {
		return rootToken, err
	}

	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&rootToken),
		autorest.ByClosing())

	return rootToken, err
}

// TODO: Add more custom actions when needed by provider resources

// TODO: The update action will need to implement operation polling as it returns an operation
