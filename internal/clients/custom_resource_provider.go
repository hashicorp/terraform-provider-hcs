package clients

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"

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

// CreateRootToken invokes the createToken Custom Resource Provider Action.
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

// FetchConsulCluster invokes the consulCluster Custom Resource Action.
func (client CustomResourceProviderClient) FetchConsulCluster(ctx context.Context, managedResourceGroupId string, clusterName string) (models.HashicorpCloudConsulamaAmaClusterResponse, error) {
	var cluster models.HashicorpCloudConsulamaAmaClusterResponse

	pathParams := map[string]interface{}{
		"resourceGroup": autorest.Encode("path", managedResourceGroupId),
		"clusterName":   autorest.Encode("path", clusterName),
	}

	const APIVersion = "2018-09-01-preview"
	queryParams := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsGet(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/{resourceGroup}/providers/Microsoft.CustomProviders/resourceProviders/public/consulClusters/{clusterName}", pathParams),
		autorest.WithQueryParameters(queryParams))

	req, err := preparer.Prepare((&http.Request{}).WithContext(ctx))
	if err != nil {
		return cluster, err
	}

	var resp *http.Response
	resp, err = client.Send(req, azure.DoRetryWithRegistration(client.Client))
	if err != nil {
		return cluster, err
	}

	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&cluster),
		autorest.ByClosing())

	return cluster, err
}

// CreateSnapshot invokes the createSnapshot Custom Resource Provider Action
func (client CustomResourceProviderClient) CreateSnapshot(ctx context.Context, managedResourceGroupID,
	resourceGroupName, snapshotName string) (models.HashicorpCloudConsulamaAmaCreateSnapshotResponse, error) {
	var snapshotResponse models.HashicorpCloudConsulamaAmaCreateSnapshotResponse

	body := models.HashicorpCloudConsulamaAmaCreateSnapshotRequest{
		Name:           snapshotName,
		ResourceGroup:  resourceGroupName,
		SubscriptionID: client.SubscriptionID,
	}

	req, err := client.customActionPreparer(ctx, managedResourceGroupID, "createSnapshot", body)
	if err != nil {
		return snapshotResponse, err
	}

	var resp *http.Response
	resp, err = client.Send(req, azure.DoRetryWithRegistration(client.Client))
	if err != nil {
		return snapshotResponse, err
	}

	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&snapshotResponse),
		autorest.ByClosing())

	return snapshotResponse, err
}

// GetSnapshot invokes the getSnapshot Custom Resource Provider Action
func (client CustomResourceProviderClient) GetSnapshot(ctx context.Context, managedResourceGroupID, resourceGroupName,
	snapshotID string) (models.HashicorpCloudConsulamaAmaGetSnapshotResponse, error) {

	var snapshotResponse models.HashicorpCloudConsulamaAmaGetSnapshotResponse

	body := models.HashicorpCloudConsulamaAmaGetSnapshotRequest{
		ResourceGroup:  resourceGroupName,
		SnapshotID:     snapshotID,
		SubscriptionID: client.SubscriptionID,
	}

	req, err := client.customActionPreparer(ctx, managedResourceGroupID, "getSnapshot", body)
	if err != nil {
		return snapshotResponse, err
	}

	var resp *http.Response
	resp, err = client.Send(req, azure.DoRetryWithRegistration(client.Client))
	if err != nil {
		return snapshotResponse, err
	}

	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&snapshotResponse),
		autorest.ByClosing())

	return snapshotResponse, err
}

// DeleteSnapshot invokes the deleteSnapshot Custom Resource Provider Action
func (client CustomResourceProviderClient) DeleteSnapshot(ctx context.Context, managedResourceGroupID, resourceGroupName,
	snapshotID string) (models.HashicorpCloudConsulamaAmaDeleteSnapshotResponse, error) {

	var snapshotResponse models.HashicorpCloudConsulamaAmaDeleteSnapshotResponse

	body := models.HashicorpCloudConsulamaAmaDeleteSnapshotRequest{
		ResourceGroup:  resourceGroupName,
		SnapshotID:     snapshotID,
		SubscriptionID: client.SubscriptionID,
	}

	req, err := client.customActionPreparer(ctx, managedResourceGroupID, "deleteSnapshot", body)
	if err != nil {
		return snapshotResponse, err
	}

	var resp *http.Response
	resp, err = client.Send(req, azure.DoRetryWithRegistration(client.Client))
	if err != nil {
		return snapshotResponse, err
	}

	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&snapshotResponse),
		autorest.ByClosing())

	return snapshotResponse, err
}

// RenameSnapshot invokes the renameSnapshot Custom Resource Provider Action
func (client CustomResourceProviderClient) RenameSnapshot(ctx context.Context, managedResourceGroupID, resourceGroupName,
	snapshotID, snapshotName string) (models.HashicorpCloudConsulamaAmaRenameSnapshotResponse, error) {

	var snapshotResponse models.HashicorpCloudConsulamaAmaRenameSnapshotResponse

	body := models.HashicorpCloudConsulamaAmaRenameSnapshotRequest{
		ResourceGroup:  resourceGroupName,
		SnapshotID:     snapshotID,
		Name:           snapshotName,
		SubscriptionID: client.SubscriptionID,
	}

	req, err := client.customActionPreparer(ctx, managedResourceGroupID, "renameSnapshot", body)
	if err != nil {
		return snapshotResponse, err
	}

	var resp *http.Response
	resp, err = client.Send(req, azure.DoRetryWithRegistration(client.Client))
	if err != nil {
		return snapshotResponse, err
	}

	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&snapshotResponse),
		autorest.ByClosing())

	return snapshotResponse, err
}

// ListUpgradeVersions invokes the listConsulUpgradeVersions Custom Resource Provider Action.
func (client CustomResourceProviderClient) ListUpgradeVersions(ctx context.Context, managedResourceGroupId string) (models.HashicorpCloudConsulamaAmaListConsulUpgradeVersionsResponse, error) {
	var upgradeVersions models.HashicorpCloudConsulamaAmaListConsulUpgradeVersionsResponse

	req, err := client.customActionPreparer(ctx, managedResourceGroupId, "listConsulUpgradeVersions", models.HashicorpCloudConsulamaAmaListConsulUpgradeVersionsRequest{
		ResourceGroup:  managedResourceGroupId,
		SubscriptionID: client.SubscriptionID,
	})
	if err != nil {
		return upgradeVersions, err
	}

	var resp *http.Response
	resp, err = client.Send(req, azure.DoRetryWithRegistration(client.Client))
	if err != nil {
		return upgradeVersions, err
	}

	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&upgradeVersions),
		autorest.ByClosing())

	return upgradeVersions, err
}

// Config invokes the config Custom Resource Provider Action.
func (client CustomResourceProviderClient) Config(ctx context.Context, managedResourceGroupId string) (models.HashicorpCloudConsulamaAmaGetConfigResponse, error) {
	var getConfigResp models.HashicorpCloudConsulamaAmaGetConfigResponse

	req, err := client.customActionPreparer(ctx, managedResourceGroupId, "config", models.HashicorpCloudConsulamaAmaGetConfigRequest{
		ResourceGroup:  managedResourceGroupId,
		SubscriptionID: client.SubscriptionID,
	})
	if err != nil {
		return getConfigResp, err
	}

	var resp *http.Response
	resp, err = client.Send(req, azure.DoRetryWithRegistration(client.Client))
	if err != nil {
		return getConfigResp, err
	}

	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&getConfigResp),
		autorest.ByClosing())

	return getConfigResp, nil
}

// UpdateCluster invokes the update Custom Resource Provider Action.
func (client CustomResourceProviderClient) UpdateCluster(ctx context.Context, managedResourceGroupID string, newConsulVersion string) (models.HashicorpCloudConsulamaAmaUpdateClusterResponse, error) {
	var updateResponse models.HashicorpCloudConsulamaAmaUpdateClusterResponse

	req, err := client.customActionPreparer(ctx, managedResourceGroupID, "update", models.HashicorpCloudConsulamaAmaUpdateClusterRequest{
		ResourceGroup:  managedResourceGroupID,
		SubscriptionID: client.SubscriptionID,
		Update: &models.HashicorpCloudConsulamaAmaClusterUpdate{
			ConsulVersion: newConsulVersion,
		},
	})
	if err != nil {
		return updateResponse, err
	}

	var resp *http.Response
	resp, err = client.Send(req, azure.DoRetryWithRegistration(client.Client))
	if err != nil {
		return updateResponse, err
	}

	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&updateResponse),
		autorest.ByClosing())

	return updateResponse, err
}

// GetFederation invokes the getFederation Custom Resource Provider Action
func (client CustomResourceProviderClient) GetFederation(ctx context.Context, managedResourceGroupID string, resourceGroupName string) (models.HashicorpCloudConsulamaAmaGetFederationResponse, error) {
	var federationResponse models.HashicorpCloudConsulamaAmaGetFederationResponse

	body := models.HashicorpCloudConsulamaAmaGetFederationRequest{
		ResourceGroup:  resourceGroupName,
		SubscriptionID: client.SubscriptionID,
	}

	req, err := client.customActionPreparer(ctx, managedResourceGroupID, "getFederation", body)
	if err != nil {
		return federationResponse, err
	}

	var resp *http.Response
	resp, err = client.Send(req, azure.DoRetryWithRegistration(client.Client))
	if err != nil {
		return federationResponse, err
	}

	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&federationResponse),
		autorest.ByClosing())

	return federationResponse, err
}

// CreateFederationToken invokes the createFederationToken Custom Resource Provider Action
func (client CustomResourceProviderClient) CreateFederationToken(ctx context.Context, managedResourceGroupID string, resourceGroupName string) (models.HashicorpCloudConsulamaAmaCreateFederationTokenResponse, error) {
	var federationTokenResponse models.HashicorpCloudConsulamaAmaCreateFederationTokenResponse

	body := models.HashicorpCloudConsulamaAmaCreateFederationTokenRequest{
		ResourceGroup:  resourceGroupName,
		SubscriptionID: client.SubscriptionID,
	}

	req, err := client.customActionPreparer(ctx, managedResourceGroupID, "createFederationToken", body)
	if err != nil {
		return federationTokenResponse, err
	}

	var resp *http.Response
	resp, err = client.Send(req, azure.DoRetryWithRegistration(client.Client))
	if err != nil {
		return federationTokenResponse, err
	}

	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&federationTokenResponse),
		autorest.ByClosing())

	return federationTokenResponse, err
}

// GetConsulConfig invokes the config Custom Resource Provider Action
func (client CustomResourceProviderClient) GetConsulConfig(ctx context.Context, managedResourceGroupID string, resourceGroupName string) (models.HashicorpCloudConsulamaAmaGetConfigResponse, error) {
	var configResponse models.HashicorpCloudConsulamaAmaGetConfigResponse

	body := models.HashicorpCloudConsulamaAmaGetConfigRequest{
		ResourceGroup:  resourceGroupName,
		SubscriptionID: client.SubscriptionID,
	}

	req, err := client.customActionPreparer(ctx, managedResourceGroupID, "config", body)
	if err != nil {
		return configResponse, err
	}

	var resp *http.Response
	resp, err = client.Send(req, azure.DoRetryWithRegistration(client.Client))
	if err != nil {
		return configResponse, err
	}

	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&configResponse),
		autorest.ByClosing())

	return configResponse, err
}

// GetOperation invokes the operation Custom Resource Provider Action
func (client CustomResourceProviderClient) GetOperation(ctx context.Context, managedResourceGroupID,
	resourceGroupName, operationID string) (models.HashicorpCloudConsulamaAmaGetOperationResponse, error) {

	var opResp models.HashicorpCloudConsulamaAmaGetOperationResponse

	body := models.HashicorpCloudConsulamaAmaGetOperationRequest{
		OperationID:    operationID,
		ResourceGroup:  resourceGroupName,
		SubscriptionID: client.SubscriptionID,
	}

	req, err := client.customActionPreparer(ctx, managedResourceGroupID, "operation", body)
	if err != nil {
		return opResp, err
	}

	var resp *http.Response
	resp, err = client.Send(req, azure.DoRetryWithRegistration(client.Client))
	if err != nil {
		return opResp, err
	}

	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusCreated),
		autorest.ByUnmarshallingJSON(&opResp),
		autorest.ByClosing())

	return opResp, err
}

// PollOperation will poll the operation Custom Resource Provider Action
// endpoint every pollInterval seconds until the operation state is DONE
// or the context cancels the request.
func (client CustomResourceProviderClient) PollOperation(ctx context.Context, operationID, managedResourceGroupID, managedAppName string,
	pollInterval int) error {

	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled")
		case <-ticker.C:
			resp, err := client.GetOperation(ctx, managedResourceGroupID, managedAppName, operationID)
			if err != nil {
				return err
			}

			if resp.Operation.State != models.HashicorpCloudConsulamaAmaOperationStateDONE {
				continue
			}

			if resp.Operation.Error != nil {
				return fmt.Errorf("an error occurred in an aysnc operation; code: %d", resp.Operation.Error.Code)
			}

			return nil
		}
	}
}
