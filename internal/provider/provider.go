package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-azure-helpers/authentication"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-hcs/internal/clients"
)

func New() func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			DataSourcesMap: map[string]*schema.Resource{
				"scaffolding_data_source": dataSourceScaffolding(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"hcs_cluster": resourceCluster(),
			},
			Schema: map[string]*schema.Schema{
				// We must support the same optional fields found in the azurerm provider schema
				// that are used for authentication to Azure. They are prefixed with azure_ below.
				"azure_subscription_id": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("ARM_SUBSCRIPTION_ID", ""),
					Description: "The Azure Subscription ID which should be used.",
				},
				"azure_client_id": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_ID", ""),
					Description: "The Azure Client ID which should be used.",
				},
				"azure_tenant_id": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("ARM_TENANT_ID", ""),
					Description: "The Azure Tenant ID which should be used.",
				},
				"azure_environment": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("ARM_ENVIRONMENT", "public"),
					Description: "The Azure Cloud Environment which should be used. Possible values are public, usgovernment, german, and china. Defaults to public.",
				},
				"azure_metadata_host": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("ARM_METADATA_HOSTNAME", ""),
					Description: "The hostname which should be used for the Azure Metadata Service.",
				},
				// Client Certificate specific fields
				"azure_client_certificate_path": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_CERTIFICATE_PATH", ""),
					Description: "The path to the Azure Client Certificate associated with the Service Principal for use when authenticating as a Service Principal using a Client Certificate.",
				},
				"azure_client_certificate_password": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_CERTIFICATE_PASSWORD", ""),
					Description: "The password associated with the Azure Client Certificate. For use when authenticating as a Service Principal using a Client Certificate",
				},
				// Client Secret specific fields
				"azure_client_secret": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_SECRET", ""),
					Description: "The Azure Client Secret which should be used. For use when authenticating as a Service Principal using a Client Secret.",
				},
				// Managed Service Identity specific fields
				"azure_use_msi": {
					Type:        schema.TypeBool,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("ARM_USE_MSI", false),
					Description: "Allowed Azure Managed Service Identity be used for Authentication.",
				},
				"azure_msi_endpoint": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("ARM_MSI_ENDPOINT", ""),
					Description: "The path to a custom endpoint for Azure Managed Service Identity - in most circumstances this should be detected automatically. ",
				},
			},
		}

		p.ConfigureContextFunc = configure(p)

		return p
	}
}

// configure returns a func that builds an authenticated Client which is used for all provider resource CRUD.
func configure(p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		builder := &authentication.Builder{
			SubscriptionID:     d.Get("azure_subscription_id").(string),
			ClientID:           d.Get("azure_client_id").(string),
			ClientSecret:       d.Get("azure_client_secret").(string),
			TenantID:           d.Get("azure_tenant_id").(string),
			Environment:        d.Get("azure_environment").(string),
			MetadataHost:       d.Get("azure_metadata_host").(string),
			MsiEndpoint:        d.Get("azure_msi_endpoint").(string),
			ClientCertPassword: d.Get("azure_client_certificate_password").(string),
			ClientCertPath:     d.Get("azure_client_certificate_path").(string),

			// Feature Toggles
			SupportsClientCertAuth:         true,
			SupportsClientSecretAuth:       true,
			SupportsManagedServiceIdentity: d.Get("azure_use_msi").(bool),
			SupportsAzureCliToken:          true,
			// TODO: Do we need to support auxiliary tenants?
			SupportsAuxiliaryTenants: false,

			// TODO: Should we keep this link to the Azure provider docs for auth?
			ClientSecretDocsLink: "https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/guides/service_principal_client_secret",
		}

		authConfig, err := builder.Build()
		if err != nil {
			return nil, diag.FromErr(fmt.Errorf("error building Azure authentication config: %s", err))
		}

		//TODO: pass provider version to user agent
		clientOptions := clients.Options{
			ProviderUserAgent: p.UserAgent("terraform-provider-hcs", ""),
			AzureAuthConfig:   authConfig,
		}

		c, err := clients.Build(ctx, clientOptions)
		if err != nil {
			return nil, diag.FromErr(fmt.Errorf("error creating HCS client: %s", err))
		}

		return c, nil
	}
}
