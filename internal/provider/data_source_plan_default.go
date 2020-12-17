package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcs/internal/clients"
	"github.com/hashicorp/terraform-provider-hcs/internal/hcsmeta"
)

// defaultPlanDefaultsTimeoutDuration is the default timeout for reading plan defaults.
var defaultPlanDefaultsTimeoutDuration = time.Minute * 5

// dataSourcePlanDefaults is the data source for the HCS plan defaults for the Azure Marketplace.
func dataSourcePlanDefaults() *schema.Resource {
	return &schema.Resource{
		Description: "The plan defaults data source is useful for accepting the Azure Marketplace Agreement for the HCS Managed Application.",
		ReadContext: dataSourcePlanDefaultsRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultPlanDefaultsTimeoutDuration,
		},
		Schema: map[string]*schema.Schema{
			// Computed outputs
			"publisher": {
				Description: "The publisher for the HCS Managed Application offer.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"offer": {
				Description: "The name of the offer for the HCS Managed Application.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"plan_name": {
				Description: "The plan name for the HCS Managed Application offer.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"plan_version": {
				Description: "The plan version for the HCS Managed Application offer.",
				Type:        schema.TypeString,
				Computed:    true,
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

	if err := d.Set("plan_name", planDefaults.Name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("plan_version", planDefaults.Version); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("publisher", "hashicorp-4665790"); err != nil {
		return diag.FromErr(err)
	}

	// Offer is set on the provider config
	if err := d.Set("offer", meta.(*clients.Client).Config.MarketPlaceProductName); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("plan_version/%s/plan_name/%s/ama_api_version/%s", planDefaults.Version, planDefaults.Name, planDefaults.ManagedAppApiVersion))

	return nil
}
