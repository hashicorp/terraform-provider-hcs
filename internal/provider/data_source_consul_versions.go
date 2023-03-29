// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcs/internal/clients"
	"github.com/hashicorp/terraform-provider-hcs/internal/consul"
)

// defaultConsulVersionsTimeoutDuration is the default timeout for reading Consul versions.
var defaultConsulVersionsTimeoutDuration = time.Minute * 5

// dataSourceConsulVersions is the data source for the Consul versions supported by HCS.
func dataSourceConsulVersions() *schema.Resource {
	return &schema.Resource{
		Description: "The Consul versions data source provides the Consul versions supported by HCS.",
		ReadContext: dataSourceConsulVersionsRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultConsulVersionsTimeoutDuration,
		},
		Schema: map[string]*schema.Schema{
			// Computed outputs
			"recommended": {
				Description: "The recommended Consul version for HCS clusters.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"available": {
				Description: "The Consul versions available on HCS.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"preview": {
				Description: "The preview versions of Consul available on HCS.",
				Type:        schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
		},
	}
}

// dataSourceConsulVersionsRead retrieves the available Consul versions from HCP and sets the schema fields
// appropriately.
func dataSourceConsulVersionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	versions, err := consul.GetAvailableHCPConsulVersions(ctx, meta.(*clients.Client).Config.HCPApiDomain)
	if err != nil {
		return diag.Errorf("unable to retrieve available Consul versions: %v", err)
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
