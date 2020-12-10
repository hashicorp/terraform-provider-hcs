package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var defaultClusterRootTokenTimeoutDuration = time.Minute * 25

// resourceClusterRootToken represents the cluster root token resource
// that is used to bootstrap the cluster's ACL system.
func resourceClusterRootToken() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterRootTokenCreate,
		ReadContext:   resourceClusterRootTokenRead,
		DeleteContext: resourceClusterRootTokenDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultClusterRootTokenTimeoutDuration,
		},
		Description: "cluster_root_token is the token used to bootstrap the cluster's ACL system",
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
			// Computed outputs
			"accessor_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kubernetes_secret": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "the root token base64 encoded in a Kubernetes secret",
			},
		},
	}
}

// resourceClusterRootTokenCreate will generate a new root token for the cluster
func resourceClusterRootTokenCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	idFromAPI := "my-id"
	d.SetId(idFromAPI)

	return diag.Errorf("not implemented")
}

// resourceClusterRootTokenRead will act as a no-op as the root token is not persisted in
// any way that it can be fetched and read
func resourceClusterRootTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("not implemented")
}

// resourceClusterRootTokenDelete will "delete" an existing token by creating a new one,
// that will not be returned, and invalidating the previous token for the cluster.
func resourceClusterRootTokenDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("not implemented")
}
