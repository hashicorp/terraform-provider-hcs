package provider

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcs/internal/clients"
	"github.com/hashicorp/terraform-provider-hcs/internal/helper"
)

// defaultClusterTimeoutDuration is the default timeout for reading the HCS cluster.
var defaultClusterTimeoutDuration = time.Minute * 5

// dataSourceCluster is the data source for an HCS Cluster.
// It has the same schema as the cluster resource, with the exception of
// consul_root_token_accessor_id and consul_root_token_secret_id.
func dataSourceCluster() *schema.Resource {
	return &schema.Resource{
		Description: "The cluster data source provides information about an existing HCS cluster.",
		ReadContext: dataSourceClusterRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultClusterTimeoutDuration,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"resource_group_name": {
				Description:      "The name of the Resource Group in which the HCS Azure Managed Application belongs.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateResourceGroupName,
			},
			"managed_application_name": {
				Description:      "The name of the HCS Azure Managed Application.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateSlugID,
			},
			// Optional inputs
			"cluster_name": {
				Description: "The name of the cluster Managed Resource. If not specified, it is defaulted to the value of `managed_application_name`.",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
			},
			// Computed outputs
			"email": {
				Description: "The contact email for the primary owner of the cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"cluster_mode": {
				Description: "The mode of the cluster ('Development' or 'Production'). Development clusters only have a single Consul server node. Production clusters deploy with a minimum of three nodes.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vnet_cidr": {
				Description: "The VNET CIDR range of the Consul cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_version": {
				Description: "The Consul version of the cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_datacenter": {
				Description: "The Consul data center name of the cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_federation_token": {
				Description: "The token used to join a federation of Consul clusters. If the cluster is not part of a federation, this field will be empty.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_external_endpoint": {
				Description: "Denotes that the cluster has an external endpoint for the Consul UI.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"location": {
				Description: "The Azure region that the cluster is deployed to.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"plan_name": {
				Description: "The name of the Azure Marketplace HCS plan for the cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"managed_resource_group_name": {
				Description: "The name of the Managed Resource Group in which the cluster resources belong.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"state": {
				Description: "The state of the cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"storage_account_name": {
				Description: "The name of the Storage Account in which cluster data is persisted.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"blob_container_name": {
				Description: "The name of the Blob Container in which cluster data is persisted.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"managed_application_id": {
				Description: "The ID of the Managed Application.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"storage_account_resource_group": {
				Description: "The name of the Storage Account's Resource Group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_automatic_upgrades": {
				Description: "Denotes that automatic Consul upgrades are enabled.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"consul_snapshot_interval": {
				Description: "The Consul snapshot interval.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_snapshot_retention": {
				Description: "The retention policy for Consul snapshots.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_config_file": {
				Description: "The cluster config encoded as a Base64 string.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_ca_file": {
				Description: "The cluster CA file encoded as a Base64 string.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_connect": {
				Description: "Denotes that Consul connect is enabled.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"consul_external_endpoint_url": {
				Description: "The public URL for the Consul UI. This will be empty if `consul_external_endpoint` is `true`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_private_endpoint_url": {
				Description: "The private URL for the Consul UI.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_cluster_id": {
				Description: "The cluster ID.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	managedAppName := d.Get("managed_application_name").(string)
	resourceGroupName := d.Get("resource_group_name").(string)

	managedApp, err := meta.(*clients.Client).ManagedApplication.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		return diag.Errorf("error fetching HCS Cluster (Resource Group Name %q) (Managed Application Name %q) (Correlation ID %q) : %+v",
			resourceGroupName,
			managedAppName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	clusterName := *managedApp.Name
	v, ok := d.GetOk("cluster_name")
	if ok {
		clusterName = v.(string)
	}

	// Fetch the cluster managed resource
	cluster, err := meta.(*clients.Client).CustomResourceProvider.FetchConsulCluster(ctx, *managedApp.ManagedResourceGroupID, clusterName)
	if err != nil {
		return diag.Errorf("error fetching HCS Cluster Managed Resource (Managed Application ID %q) (Cluster Name %q) (Correlation ID %q): %+v",
			*managedApp.ID,
			clusterName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	// Fetch the managed VNet
	managedResourceGroupName, err := helper.ParseResourceGroupNameFromID(*managedApp.ManagedResourceGroupID)
	if err != nil {
		return diag.FromErr(err)
	}

	// VNet name has a '-vnet' suffix that is not saved on the cluster properties
	vNetName := strings.TrimSuffix(cluster.Properties.VnetName, "-vnet") + "-vnet"
	vNet, err := meta.(*clients.Client).VNet.Get(ctx, managedResourceGroupName, vNetName, "")
	if err != nil {
		return diag.Errorf("error fetching VNet for HCS Cluster (Managed Application ID %q) (Managed Resource Group Name %q) (VNet Name %q) (Correlation ID %q): %+v",
			*managedApp.ID,
			managedResourceGroupName,
			vNetName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	d.SetId(*managedApp.ID)

	return setClusterData(d, managedApp, cluster, vNet)
}
