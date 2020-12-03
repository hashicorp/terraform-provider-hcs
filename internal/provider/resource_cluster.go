package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-provider-hcs/internal/timeouts"

	"github.com/hashicorp/terraform-provider-hcs/internal/clients"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var createUpdateDeleteTimeoutDuration = time.Minute * 25

func resourceCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		UpdateContext: resourceClusterUpdate,
		DeleteContext: resourceClusterDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: &createUpdateDeleteTimeoutDuration,
			Update: &createUpdateDeleteTimeoutDuration,
			Delete: &createUpdateDeleteTimeoutDuration,
		},
		Schema: map[string]*schema.Schema{
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
			"email": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateStringNotEmpty,
			},
			"cluster_mode": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateDiagFunc: validateStringInSlice([]string{
					"Development",
					"Production",
				}, true),
			},
			"cluster_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				// TODO: can we validate optional fields
				ValidateDiagFunc: validateSlugID,
			},
			"vnet_cidr": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          "172.25.16.0/24",
				ValidateDiagFunc: validateCIDR,
			},
			"consul": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validateConsulVersion,
							Computed:         true,
						},
						"datacenter": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validateSlugID,
							ForceNew:         true,
						},
						"federation_token": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"external_endpoint": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
							ForceNew: true,
						},
					},
				},
			},
			"location": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				// TODO: validate location the same way azurerm does
			},
			"plan_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"managed_resource_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctx, cancel := timeouts.ForCreateUpdate(ctx, d)
	defer cancel()

	managedAppName := d.Get("managed_application_name").(string)
	resourceGroupName := d.Get("resource_group_name").(string)

	_, _ = meta.(*clients.Client).CustomResourceProvider.CreateRootToken(ctx, "INSERT_MRG_HERE")

	managedAppClient := meta.(*clients.Client).ManagedApplication
	existing, err := managedAppClient.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		if existing.Response.StatusCode != 404 {
			return diag.Errorf("failed to check for present of existing Managed Application Name %q (Resource Group %q): %+v", managedAppName, resourceGroupName, err)
		}
	}
	if existing.ID != nil && *existing.ID != "" {
		msg := "a resource with the ID %q already exists - to be managed via Terraform this resource needs to be imported into the State. Please see the resource documentation for hcs_cluster for more information"
		return diag.Errorf(msg, *existing.ID)
	}

	// TODO: set defaults for values that are dependent on side effects / other schema values
	//  consul_version, datacenter, plan_name, managed_resource_group_name, location
	//clusterName := managedAppName
	//v, ok := d.GetOk("cluster_name")
	//if ok {
	//	clusterName = v.(string)
	//}
	//
	//managedResourceGroupName := managedAppName
	//v, ok = d.GetOk("cluster_name")
	//if ok {
	//	clusterName = v.(string)
	//}
	//
	//// TODO set consul defaults
	//datacenter := managedAppName
	//v, ok = d.GetOk("datacenter")
	//if ok {
	//	datacenter = v.(string)
	//}
	//// TODO fetch plan defaults
	//params := managedapplications.Application{
	//	ApplicationProperties: nil,
	//	Plan:                  nil,
	//	Kind:                  nil,
	//	Identity:              nil,
	//	ManagedBy:             nil,
	//	Sku:                   nil,
	//	ID:                    nil,
	//	Name:                  nil,
	//	Type:                  nil,
	//	Location:              nil,
	//	Tags:                  nil,
	//}
	//future, err := managedAppClient.CreateOrUpdate(ctx, resourceGroupName, managedAppName, params)
	//if err != nil {
	//	return diag.Errorf("failed to create HCS Cluster (Managed Application %q) (Resource Group %q): %+v", managedAppName, resourceGroupName, err)
	//}
	//if err = future.WaitForCompletionRef(ctx, managedAppClient.Client); err != nil {
	//	return diag.Errorf("failed to wait for creation of HCS Cluster (Managed Application %q) (Resource Group %q): %+v", managedAppName, resourceGroupName, err)
	//}
	//
	//app, err := managedAppClient.Get(ctx, resourceGroupName, managedAppName)
	//if err != nil {
	//	return diag.Errorf("failed to retrieve HCS Cluster (Managed Application %q) (Resource Group %q): %+v", managedAppName, resourceGroupName, err)
	//}
	//if app.ID == nil || *app.ID == "" {
	//	return diag.Errorf("cannot read HCS Cluster (Managed Application %q) (Resource Group %q) ID", managedAppName, resourceGroupName)
	//}
	//
	//// Create a token
	//crpClient := meta.(*clients.Client).CustomResourceProvider
	//rootTokenResp, err := crpClient.CreateRootToken(ctx, *app.ApplicationProperties.ManagedResourceGroupID)
	//if err != nil {
	//	return diag.Errorf("failed to create HCS Cluster root token (Managed Application %q) (Resource Group %q) ID", managedAppName, resourceGroupName)
	//}
	//
	//d.SetId(*app.ID)
	//d.Set("outputs")

	return diag.Errorf("not implemented")
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("not implemented")
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("not implemented")
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("not implemented")
}
