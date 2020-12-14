package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcs/internal/clients"
	"github.com/hashicorp/terraform-provider-hcs/internal/hcsmeta"
)

// dataSourcePlanDefaults is the data source for the HCS plan defaults for the Azure Marketplace.
func dataSourcePlanDefaults() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePlanDefaultsRead,
		Schema: map[string]*schema.Schema{
			// Computed outputs
			"publisher": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"offer": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"plan_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"plan_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// dataSourcePlanDefaultsRead retrieves the HCS Meta plan defaults and sets the HCS plan defaults for
// the Azure marketplace.
func dataSourcePlanDefaultsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	planDefaults, err := hcsmeta.GetPlanDefaults(ctx)
	if err != nil {
		return diag.Errorf("unable to retrieve HCS Meta plan defaults: %+v", err)
	}

	err = d.Set("plan_name", planDefaults.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("plan_version", planDefaults.Version)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("publisher", "hashicorp-4665790")
	if err != nil {
		return diag.FromErr(err)
	}

	// Offer is set on the provider config
	err = d.Set("offer", meta.(*clients.Client).Config.MarketPlaceProductName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("plan_version/%s/plan_name/%s/ama_api_version/%s", planDefaults.Version, planDefaults.Name, planDefaults.ManagedAppApiVersion))

	return nil
}
