package clients

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/go-azure-helpers/authentication"
	"github.com/hashicorp/go-azure-helpers/sender"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-07-01/managedapplications"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2020-06-01/resources"
)

var (
	// senderProviderName is the friendly name of the provider and is output in Autorest sender request/response logs.
	senderProviderName = "HashicorpConsulService"
)

// Options are the options passed to the client.
type Options struct {
	// ProviderUserAgent is the User Agent used for HTTP requests which contains the provider name and version.
	ProviderUserAgent string

	// AzureAuthConfig is the configuration used to create an authenticated Azure client.
	AzureAuthConfig *authentication.Config
}

// Client is used by the provider to make authenticated HTTP requests to Azure.
type Client struct {
	// Account contains Azure RM account information.
	Account *AzureResourceManagerAccount

	// ManagedApplication is the client used for Azure Managed Application CRUD.
	ManagedApplication *managedapplications.ApplicationsClient

	// ResourceGroup is the client used for Azure Resource Group CRUD
	ResourceGroup *resources.GroupsClient

	// CustomResourceProvider is the client used for HCS Custom Resource Provider actions.
	CustomResourceProvider *CustomResourceProviderClient
}

// Build constructs a Client which is used by the provider to make authenticated HTTP requests to Azure.
// Adapted from the azurerm provider's clients.Build
// https://github.com/terraform-providers/terraform-provider-azurerm/blob/8f32ad645888ee00a24ad7c739a8703222e13913/azurerm/internal/clients/builder.go#L38
func Build(ctx context.Context, options Options) (*Client, error) {
	env, err := authentication.AzureEnvironmentByNameFromEndpoint(ctx, options.AzureAuthConfig.MetadataHost, options.AzureAuthConfig.Environment)
	if err != nil {
		return nil, err
	}

	account, err := NewAzureResourceManagerAccount(ctx, *options.AzureAuthConfig, *env)
	if err != nil {
		return nil, fmt.Errorf("error building account: %+v", err)
	}

	client := Client{
		Account: account,
	}

	oauthConfig, err := options.AzureAuthConfig.BuildOAuthConfig(env.ActiveDirectoryEndpoint)
	if err != nil {
		return nil, err
	}
	if oauthConfig == nil {
		return nil, fmt.Errorf("unable to configure OAuthConfig for tenant %s", options.AzureAuthConfig.TenantID)
	}

	send := sender.BuildSender(senderProviderName)
	auth, err := options.AzureAuthConfig.GetAuthorizationToken(send, oauthConfig, env.TokenAudience)
	if err != nil {
		return nil, err
	}

	// Prevent rate limited requests to be counted against the request retry count.
	autorest.Count429AsRetry = false

	managedAppClient := managedapplications.NewApplicationsClientWithBaseURI(env.ResourceManagerEndpoint, options.AzureAuthConfig.SubscriptionID)
	configureAutoRestClient(&managedAppClient.Client, auth, options.ProviderUserAgent)
	client.ManagedApplication = &managedAppClient

	resourceGroupClient := resources.NewGroupsClient(options.AzureAuthConfig.SubscriptionID)
	configureAutoRestClient(&resourceGroupClient.Client, auth, options.ProviderUserAgent)
	client.ResourceGroup = &resourceGroupClient

	customResourceProviderClient := NewCustomResourceProviderClientWithBaseURI(env.ResourceManagerEndpoint, options.AzureAuthConfig.SubscriptionID)
	configureAutoRestClient(&customResourceProviderClient.Client, auth, options.ProviderUserAgent)
	client.CustomResourceProvider = &customResourceProviderClient

	return &client, nil
}

// configureAutoRestClient is used to configure an Azure Autorest client with the appropriate User Agent,
// authorizer, and correlation id etc.
func configureAutoRestClient(c *autorest.Client, authorizer autorest.Authorizer, providerUserAgent string) {
	c.UserAgent = strings.TrimSpace(fmt.Sprintf("%s %s", c.UserAgent, providerUserAgent))

	c.Authorizer = authorizer
	c.Sender = sender.BuildSender(senderProviderName)
	c.RequestInspector = withCorrelationRequestID(correlationRequestID())
}
