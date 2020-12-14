package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
func dataSourceAgentHelmConfigRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
