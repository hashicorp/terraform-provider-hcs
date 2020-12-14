package provider

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcs/internal/clients"
)

// dataSourceAgentHelmConfig is the data source for the agent Helm
// config for an HCS cluster.
func dataSourceAgentHelmConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAgentHelmConfigRead,
		Schema: map[string]*schema.Schema{
			// Required inputs
			"resource_group_name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateResourceGroupName,
			},
			"managed_application_name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"aks_cluster_name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateStringNotEmpty,
			},
			// Optional
			"aks_resource_group": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateStringNotEmpty,
			},
			// Computed outputs
			"config": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// dataSourceAgentHelmConfigRead is the func to implement reading of the
// agent Helm config for an HCS cluster.
func dataSourceAgentHelmConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resourceGroupName := d.Get("resource_group_name").(string)
	managedAppName := d.Get("managed_application_name").(string)

	managedAppClient := meta.(*clients.Client).ManagedApplication
	app, err := managedAppClient.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		return diag.Errorf("failed to check for presence of existing HCS Cluster (Managed Application %q) (Resource Group %q): %+v", managedAppName, resourceGroupName, err)
	}
	if app.Response.StatusCode == 404 {
		// No managed application exists, so returning an error stating as such
		return diag.Errorf("[ERROR] no HCS Cluster found for (Managed Application %q) (Resource Group %q).", managedAppName, resourceGroupName)
	}

	managedAppManagedResourceGroupID := *app.ManagedResourceGroupID

	crpClient := meta.(*clients.Client).CustomResourceProvider

	resp, err := crpClient.Config(ctx, managedAppManagedResourceGroupID)
	if err != nil {
		return diag.Errorf("failed to get config for managed app: %+v", err)
	}

	log.Printf("[ERROR] Client config: %q", resp.ClientConfig)

	return nil
}
