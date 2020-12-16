package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcs/internal/clients"
)

// dataSourceFederationToken represents a federation token for an HCS Cluster.
func dataSourceFederationToken() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFederationTokenRead,
		Schema: map[string]*schema.Schema{
			// Required inputs
			"resource_group_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateResourceGroupName,
			},
			"managed_application_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			// Computed output
			"token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
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
		return diag.Errorf("error fetching HCS Cluster to be used as federation primary (Managed Application %q) (Resource Group %q) (Correlation ID %q): %+v",
			managedAppName,
			resourceGroupName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	federationTokenResponse, err := meta.(*clients.Client).CustomResourceProvider.CreateFederationToken(ctx, *managedApp.ManagedResourceGroupID, resourceGroupName)
	if err != nil {
		return diag.Errorf("error fetching a federation token for primary cluster (Managed Application %q) (Resource Group %q) (Correlation ID %q): %+v",
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

	return nil
}
