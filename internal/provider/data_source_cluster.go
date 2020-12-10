package provider

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-provider-hcs/internal/clients"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dataSourceCluster is the data source for an HCS Cluster.
// It has the same schema as the cluster resource, with the exception of
// consul_root_token_accessor_id and consul_root_token_secret_id.
func dataSourceCluster() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceClusterRead,
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
			// Optional inputs
			"cluster_name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			// Computed outputs
			"email": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vnet_cidr": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"consul_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"consul_datacenter": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"consul_federation_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"consul_external_endpoint": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"plan_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"managed_resource_group_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_account_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"blob_container_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"managed_application_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_account_resource_group": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"consul_automatic_upgrades": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"consul_snapshot_interval": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"consul_snapshot_retention": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"consul_config_file": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"consul_ca_file": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"consul_connect": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"consul_external_endpoint_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"consul_private_endpoint_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"consul_cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	managedAppName := d.Get("managed_application_name").(string)
	resourceGroupName := d.Get("resource_group_name").(string)

	managedApp, err := meta.(*clients.Client).ManagedApplication.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		if managedApp.Response.StatusCode == 404 {
			log.Printf("[INFO] HCS Cluster %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return diag.Errorf("error fetching HCS Cluster (Resource Group Name %q) (Managed Application Name %q) : %+v", resourceGroupName, managedAppName, err)
	}

	clusterName := *managedApp.Name
	v, ok := d.GetOk("cluster_name")
	if ok {
		clusterName = v.(string)
	}

	// Fetch the cluster managed resource
	cluster, err := meta.(*clients.Client).CustomResourceProvider.FetchConsulCluster(ctx, *managedApp.ManagedResourceGroupID, clusterName)
	if err != nil {
		return diag.Errorf("error fetching HCS Cluster (Managed Application ID %q) (Cluster Name %q): %+v", *managedApp.ID, clusterName, err)
	}

	d.SetId(*managedApp.ID)

	return setClusterData(d, managedApp, cluster)
}
