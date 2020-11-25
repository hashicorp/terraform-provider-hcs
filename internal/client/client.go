package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/go-azure-helpers/authentication"
	"github.com/hashicorp/go-azure-helpers/sender"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-07-01/managedapplications"
)

type Options struct {
	ProviderUserAgent string

	AzureAuthConfig *authentication.Config
}

type Client struct {
	// StopContext is used for propagating control from Terraform Core (e.g. Ctrl/Cmd+C)
	StopContext context.Context

	Account *AzureResourceManagerAccount

	ManagedApplication *managedapplications.ApplicationsClient
}

func NewClient(ctx context.Context, options Options) (*Client, error) {
	client := Client{
		StopContext: ctx,
	}

	env, err := authentication.AzureEnvironmentByNameFromEndpoint(ctx, options.AzureAuthConfig.MetadataHost, options.AzureAuthConfig.Environment)
	if err != nil {
		return nil, err
	}

	account, err := NewAzureResourceManagerAccount(ctx, *options.AzureAuthConfig, *env)
	if err != nil {
		return nil, fmt.Errorf("error building account: %+v", err)
	}
	client.Account = account

	oauthConfig, err := options.AzureAuthConfig.BuildOAuthConfig(env.ActiveDirectoryEndpoint)
	if err != nil {
		return nil, err
	}
	if oauthConfig == nil {
		return nil, fmt.Errorf("unable to configure OAuthConfig for tenant %s", options.AzureAuthConfig.TenantID)
	}

	send := sender.BuildSender("HCS")
	auth, err := options.AzureAuthConfig.GetAuthorizationToken(send, oauthConfig, env.TokenAudience)
	if err != nil {
		return nil, err
	}

	autorest.Count429AsRetry = false
	managedAppClient := managedapplications.NewApplicationsClientWithBaseURI(env.ResourceManagerEndpoint, options.AzureAuthConfig.SubscriptionID)
	configureAutoRestClient(&managedAppClient.Client, auth, options.ProviderUserAgent)

	// TODO: Wire up a management client to make Custom Resource Provider requests

	return &client, nil
}

func configureAutoRestClient(c *autorest.Client, authorizer autorest.Authorizer, providerUserAgent string) {
	c.UserAgent = strings.TrimSpace(fmt.Sprintf("%s %s", c.UserAgent, providerUserAgent))

	c.Authorizer = authorizer
	c.Sender = sender.BuildSender("HCS")
	c.RequestInspector = withCorrelationRequestID(correlationRequestID())
}
