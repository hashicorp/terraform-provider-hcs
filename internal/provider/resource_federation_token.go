package provider

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcs/internal/clients"
)

// resourceFederationToken represents a federation token for an HCS Cluster.
func resourceFederationToken() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFederationTokenCreate,
		ReadContext:   resourceFederationTokenRead,
		DeleteContext: resourceFederationTokenDelete,
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

func resourceFederationTokenCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	managedAppName := d.Get("managed_application_name").(string)
	resourceGroupName := d.Get("resource_group_name").(string)

	managedApp, err := meta.(*clients.Client).ManagedApplication.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		return diag.Errorf("failed to fetch HCS Cluster to be used as federation primary (Managed Application %q) (Resource Group %q): %+v", managedAppName, resourceGroupName, err)
	}

	// TODO: Does setting the resource id to the AMA id prevent us from
	// being able to create another token for the same cluster without deleting this one?
	// If so, that is good. If not, we should probably generate a UUID.
	d.SetId(*managedApp.ID + "/federation-token")

	federationTokenResponse, err := meta.(*clients.Client).CustomResourceProvider.CreateFederationToken(ctx, *managedApp.ManagedResourceGroupID, resourceGroupName)
	if err != nil {
		// Remove this if ID is set after success
		d.SetId("")
		return diag.Errorf("failed to create federation token for primary cluster (Managed Application %q) (Resource Group %q): %+v", managedAppName, resourceGroupName, err)
	}

	err = d.Set("token", federationTokenResponse.FederationToken)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceFederationTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resourceGroupName := d.Get("resource_group_name").(string)
	managedAppName := d.Get("managed_application_name").(string)

	app, err := meta.(*clients.Client).ManagedApplication.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		return diag.Errorf("failed to check for presence of existing HCS Cluster (Managed Application %q) (Resource Group %q): %+v", managedAppName, resourceGroupName, err)
	}
	if app.Response.StatusCode == 404 {
		// The primary cluster of the federation (managed application) no longer exists, so this federation token should be removed from state
		log.Printf("[INFO] the primary cluster for this federation token was not found for (Managed Application %q) (Resource Group %q)", managedAppName, resourceGroupName)
		d.SetId("")
		return nil
	}
	return nil
}

func resourceFederationTokenDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// TODO: should this be a NoOp or should we create a new federation token that is not stored?
	log.Print("[DEBUG] HCS federation token delete is a NoOp")
	return nil
}
