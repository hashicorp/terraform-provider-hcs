package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-provider-hcs/internal/hcsmeta"

	"github.com/hashicorp/terraform-provider-hcs/internal/consul"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-07-01/managedapplications"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcs/internal/clients"
	"github.com/hashicorp/terraform-provider-hcs/internal/timeouts"
	"github.com/hashicorp/terraform-provider-hcs/utils"
)

var createUpdateDeleteTimeoutDuration = time.Minute * 25

type managedAppParamValue struct {
	Value interface{} `json:"value"`
}

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
			// Optional inputs
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
			"consul_version": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: validateConsulVersion,
			},
			"consul_datacenter": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateSlugID,
				ForceNew:         true,
			},
			"consul_federation_token": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"consul_external_endpoint": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
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
			// Computed outputs
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
			"consul_root_token_accessor_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"consul_root_token_secret_id": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctx, cancel := timeouts.ForCreateUpdate(ctx, d)
	defer cancel()

	managedAppName := d.Get("managed_application_name").(string)
	resourceGroupName := d.Get("resource_group_name").(string)

	managedAppClient := meta.(*clients.Client).ManagedApplication

	// Ensure a managed app with the same name does not exist in this resource group
	existingCluster, err := managedAppClient.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		if existingCluster.Response.StatusCode != 404 {
			return diag.Errorf("failed to check for presence of existing HCS Cluster (Managed Application %q) (Resource Group %q): %+v", managedAppName, resourceGroupName, err)
		}
	}
	if existingCluster.ID != nil && *existingCluster.ID != "" {
		return diag.Errorf("a resource with the ID %q already exists - to be managed via Terraform this resource needs to be imported into the State. Please see the resource documentation for hcs_cluster for more information", *existingCluster.ID)
	}

	// Fetch resource group
	resourceGroup, err := meta.(*clients.Client).ResourceGroup.Get(ctx, resourceGroupName)
	if err != nil {
		return diag.Errorf("failed to fetch resource group (Resource Group %q):  %+v", resourceGroupName, err)
	}

	location := resourceGroup.Location
	v, ok := d.GetOk("location")
	if ok {
		location = utils.String(strings.ReplaceAll(strings.ToLower(v.(string)), " ", ""))
	}

	clusterName := managedAppName
	v, ok = d.GetOk("cluster_name")
	if ok {
		clusterName = v.(string)
	}

	managedResourceGroupId := *resourceGroup.ID + "-mrg-" + managedAppName
	v, ok = d.GetOk("managed_resource_group_name")
	if ok {
		managedResourceGroupId = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", meta.(*clients.Client).Account.SubscriptionId, v.(string))
	}

	// Consul defaults
	dataCenter := managedAppName
	v, ok = d.GetOk("consul_datacenter")
	if ok {
		dataCenter = v.(string)
	}

	externalEndpoint := "disabled"
	if d.Get("consul_external_endpoint").(bool) {
		externalEndpoint = "enabled"
	}

	availableConsulVersions, err := consul.GetAvailableHCPConsulVersions(ctx, meta.(*clients.Client).Config.HCPApiDomain)
	if err != nil || availableConsulVersions == nil {
		return diag.Errorf("failed to get available HCP Consul versions: %+v", err)
	}
	consulVersion := consul.RecommendedVersion(availableConsulVersions)
	v, ok = d.GetOk("consul_version")
	if ok {
		consulVersion = consul.NormalizeVersion(v.(string))
	}
	if !consul.IsValidVersion(consulVersion, availableConsulVersions) {
		return diag.Errorf("specified Consul version (%s) is unavailable; must be one of: %+v", availableConsulVersions)
	}

	var federationToken string
	v, ok = d.GetOk("consul_federation_token")
	if ok {
		federationToken = v.(string)
	}

	// Azure Marketplace Plan
	planDefaults, err := hcsmeta.GetPlanDefaults(ctx)
	if err != nil {
		return diag.Errorf("unable to retrieve HCS Azure Marketplace plan defaults: %+v", err)
	}

	planName := planDefaults.Name
	v, ok = d.GetOk("plan_name")
	if ok {
		planName = v.(string)
	}

	plan := managedapplications.Plan{
		Name:      utils.String(planName),
		Version:   utils.String(planDefaults.Version),
		Product:   utils.String(meta.(*clients.Client).Config.MarketPlaceProductName),
		Publisher: utils.String("hashicorp-4665790"),
	}

	hcsAMAParams := map[string]managedAppParamValue{
		"clusterName": {
			Value: clusterName,
		},
		"consulDataCenter": {
			Value: dataCenter,
		},
		"consulVnetCidr": {
			Value: d.Get("vnet_cidr").(string),
		},
		"email": {
			Value: d.Get("email").(string),
		},
		"externalEndpoint": {
			Value: externalEndpoint,
		},
		"initialConsulVersion": {
			Value: consulVersion,
		},
	}

	if federationToken != "" {
		hcsAMAParams["federationToken"] = managedAppParamValue{
			Value: federationToken,
		}
	}

	params := managedapplications.Application{
		ApplicationProperties: &managedapplications.ApplicationProperties{
			ManagedResourceGroupID: utils.String(managedResourceGroupId),
			Parameters:             hcsAMAParams,
		},
		Plan:     &plan,
		Kind:     utils.String("MarketPlace"),
		Location: location,
	}
	future, err := managedAppClient.CreateOrUpdate(ctx, resourceGroupName, managedAppName, params)
	if err != nil {
		return diag.Errorf("failed to create HCS Cluster (Managed Application %q) (Resource Group %q): %+v", managedAppName, resourceGroupName, err)
	}
	if err = future.WaitForCompletionRef(ctx, managedAppClient.Client); err != nil {
		return diag.Errorf("failed to wait for creation of HCS Cluster (Managed Application %q) (Resource Group %q): %+v", managedAppName, resourceGroupName, err)
	}

	app, err := managedAppClient.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		return diag.Errorf("failed to retrieve HCS Cluster (Managed Application %q) (Resource Group %q): %+v", managedAppName, resourceGroupName, err)
	}
	if app.ID == nil || *app.ID == "" {
		return diag.Errorf("cannot read HCS Cluster (Managed Application %q) (Resource Group %q) ID", managedAppName, resourceGroupName)
	}
	d.SetId(*app.ID)

	// Create a token
	crpClient := meta.(*clients.Client).CustomResourceProvider
	rootTokenResp, err := crpClient.CreateRootToken(ctx, *app.ApplicationProperties.ManagedResourceGroupID)
	if err != nil {
		return diag.Errorf("failed to create HCS Cluster root token (Managed Application %q) (Resource Group %q) ID", managedAppName, resourceGroupName)
	}

	// Only set root token keys after create
	err = d.Set("consul_root_token_accessor_id", rootTokenResp.MasterToken.AccessorID)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("consul_root_token_secret_id", rootTokenResp.MasterToken.SecretID)
	if err != nil {
		return diag.FromErr(err)
	}

	return setClusterResourceData(d, app)
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

func setClusterResourceData(d *schema.ResourceData, managedApp managedapplications.Application) diag.Diagnostics {
	return diag.Errorf("not implemented")
}
