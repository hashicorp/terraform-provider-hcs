package client

import (
	"context"
	"net/http"

	"github.com/Azure/go-autorest/autorest/azure"

	"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/terraform-provider-hcs/internal/client/hcs-ama-api-spec/models"
)

type CustomResourceProviderClient struct {
	autorest.Client
	BaseURI        string
	SubscriptionID string
}

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
		"resourceGroup": autorest.Encode("path", managedResourceGroupId),
		"action":        autorest.Encode("path", action),
	}

	const APIVersion = "2018-09-01-preview"
	queryParams := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsPost(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/{resourceGroup}/providers/Microsoft.CustomProviders/resourceProviders/public/{action}", pathParams),
		autorest.WithJSON(body),
		autorest.WithQueryParameters(queryParams))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// CreateRootToken invokes the createToken Custom Resource Action.
func (client CustomResourceProviderClient) CreateRootToken(ctx context.Context, managedResourceGroupId string, body models.HashicorpCloudConsulamaAmaCreateTokenRequest) (models.HashicorpCloudConsulamaAmaCreateTokenResponse, error) {
	var rootToken models.HashicorpCloudConsulamaAmaCreateTokenResponse

	req, err := client.customActionPreparer(ctx, managedResourceGroupId, "createToken", body)
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
