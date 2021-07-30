package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcs/internal/clients"
)

// defaultFederationTokenTimeoutDuration is the default timeout for reading a federation token.
var defaultFederationTokenTimeoutDuration = time.Minute * 5

// dataSourceFederationToken represents a federation token for an HCS Cluster.
func dataSourceFederationToken() *schema.Resource {
	return &schema.Resource{
		Description: "The federation token data source can be used during HCS cluster creation to join the cluster to a federation.",
		ReadContext: dataSourceFederationTokenRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultFederationTokenTimeoutDuration,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"resource_group_name": {
				Description:      "The name of the Resource Group in which the HCS Azure Managed Application belongs.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateResourceGroupName,
			},
			"managed_application_name": {
				Description:      "The name of the HCS Azure Managed Application.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			// Computed output
			"token": {
				Description: "The federation token.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

// dataSourceFederationTokenRead gets a new federation token for the HCS cluster.
// Since federation tokens are not persisted in HCS, we generate a new one for each
// data source read.
func dataSourceFederationTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	managedAppName := d.Get("managed_application_name").(string)
	resourceGroupName := d.Get("resource_group_name").(string)

	managedApp, err := meta.(*clients.Client).ManagedApplication.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		return diag.Errorf("unable to fetch HCS cluster to be used as primary federation cluster (Managed Application %q) (Resource Group %q) (Correlation ID %q): %v",
			managedAppName,
			resourceGroupName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	// An error here denotes that the cluster is not part of a federation
	federationResponse, err := meta.(*clients.Client).CustomResourceProvider.GetFederation(ctx, *managedApp.ManagedResourceGroupID, d.Get("resource_group_name").(string))
	// Ensure the cluster is the primary in the federation
	if err == nil && isClusterPrimaryInFederation(*managedApp.Name, resourceGroupName, federationResponse) {
		federationTokenResponse, err := meta.(*clients.Client).CustomResourceProvider.CreateFederationToken(ctx, *managedApp.ManagedResourceGroupID, resourceGroupName)
		if err != nil {
			return diag.Errorf("unable to fetch a federation token for primary cluster (Managed Application %q) (Resource Group %q) (Correlation ID %q): %v",
				managedAppName,
				resourceGroupName,
				meta.(*clients.Client).CorrelationRequestID,
				err,
			)
		}

		err = d.Set("token", federationTokenResponse.FederationToken)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(*managedApp.ID + "/federation-token")
	}

	return nil
}
