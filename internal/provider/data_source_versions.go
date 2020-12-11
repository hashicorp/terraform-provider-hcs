package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-provider-hcs/internal/consul"

	"github.com/hashicorp/terraform-provider-hcs/internal/clients"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceVersions is the data source for the Consul versions supported by HCS.
func dataSourceVersions() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVersionsRead,
		Schema: map[string]*schema.Schema{
			// Computed outputs
			"recommended": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"available": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"preview": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
		},
	}
}

// dataSourceVersionsRead retrieves the available Consul versions from HCP and sets the schema fields
// appropriately.
func dataSourceVersionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	versions, err := consul.GetAvailableHCPConsulVersions(ctx, meta.(*clients.Client).Config.HCPApiDomain)
	if err != nil {
		return diag.Errorf("unable to retrieve available Consul versions: %+v", err)
	}

	var recommendedVersion string
	availableVersions := make([]string, 0)
	previewVersions := make([]string, 0)

	for _, v := range versions {
		switch v.Status {
		case "RECOMMENDED":
			recommendedVersion = v.Version
			availableVersions = append(availableVersions, v.Version)
		case "AVAILABLE":
			availableVersions = append(availableVersions, v.Version)
		case "PREVIEW":
			previewVersions = append(previewVersions, v.Version)
		}
	}

	err = d.Set("recommended", recommendedVersion)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("available", availableVersions)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("preview", previewVersions)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("recommended/%s/available_len/%d/preview_len/%d", recommendedVersion, len(availableVersions), len(previewVersions)))

	return nil
}
