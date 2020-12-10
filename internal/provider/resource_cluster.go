package provider

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-07-01/managedapplications"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcs/internal/clients"
	"github.com/hashicorp/terraform-provider-hcs/internal/clients/hcs-ama-api-spec/models"
	"github.com/hashicorp/terraform-provider-hcs/internal/consul"
	"github.com/hashicorp/terraform-provider-hcs/internal/hcsmeta"
	"github.com/hashicorp/terraform-provider-hcs/internal/helper"
	"github.com/hashicorp/terraform-provider-hcs/utils"
)

var createUpdateDeleteTimeoutDuration = time.Minute * 25

// managedAppParamValue is the container struct for passing AMA values on creation/update.
type managedAppParamValue struct {
	// Value is the value of the AMA param
	Value interface{} `json:"value"`
}

// resourceCluster represents an HCS Cluster resource.
// Most of the CRUD involves the Azure Managed Application and Custom Resource Provider actions.
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
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					return strings.ToLower(old) == strings.ToLower(new)
				},
			},
			// Optional inputs
			"cluster_name": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Computed:         true,
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
				ValidateDiagFunc: validateSemVer,
			},
			"consul_datacenter": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateSlugID,
				ForceNew:         true,
				Computed:         true,
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
				Computed: true,
				// TODO: validate location the same way azurerm does
			},
			"plan_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ValidateDiagFunc: validateStringInSlice([]string{
					"on-demand-v2",
					"annual",
				}, false),
			},
			"managed_resource_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
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
	supportedRegions, err := hcsmeta.GetSupportedRegions(ctx)
	if err != nil {
		return diag.Errorf("unable to retrieve supported HCS regions: %+v", err)
	}
	if !hcsmeta.RegionIsSupported(*location, supportedRegions) {
		return diag.Errorf("unsupported location: %s; expected location to be one of %+v", *location, supportedRegions)
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
		return diag.Errorf("specified Consul version (%s) is unavailable; must be one of: %+v", consulVersion, availableConsulVersions)
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

	clusterName := managedAppName
	v, ok = d.GetOk("cluster_name")
	if ok {
		clusterName = v.(string)
	}

	managedResourceGroupId := fmt.Sprintf("%s-mrg-%s", *resourceGroup.ID, managedAppName)
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

	var federationToken string
	v, ok = d.GetOk("consul_federation_token")
	if ok {
		federationToken = v.(string)
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

	// We need to set the cluster name to be able to fetch the cluster on read
	err = d.Set("cluster_name", clusterName)

	d.SetId(*app.ID)

	rootTokenResp, err := meta.(*clients.Client).CustomResourceProvider.CreateRootToken(ctx, *app.ApplicationProperties.ManagedResourceGroupID)
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

	return resourceClusterRead(ctx, d, meta)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Fetch the managed app
	managedAppID := d.Id()
	managedApp, err := meta.(*clients.Client).ManagedApplication.GetByID(ctx, managedAppID)
	if err != nil {
		if managedApp.Response.StatusCode == 404 {
			log.Printf("[INFO] HCS Cluster %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return diag.Errorf("error fetching HCS Cluster (Managed Application ID %q) : %+v", managedAppID, err)
	}

	clusterName := *managedApp.Name
	v, ok := d.GetOk("cluster_name")
	if ok {
		clusterName = v.(string)
	}

	// Fetch the cluster managed resource
	cluster, err := meta.(*clients.Client).CustomResourceProvider.FetchConsulCluster(ctx, *managedApp.ManagedResourceGroupID, clusterName)
	if err != nil {
		return diag.Errorf("error fetching HCS Cluster (Managed Application ID %q) (Cluster Name %q): %+v", managedAppID, clusterName, err)
	}

	return setClusterResourceData(d, managedApp, cluster)
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Fetch the managed app
	managedAppID := d.Id()
	managedApp, err := meta.(*clients.Client).ManagedApplication.GetByID(ctx, managedAppID)
	if err != nil {
		if managedApp.Response.StatusCode == 404 {
			log.Printf("[INFO] HCS Cluster %q does not exist - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return diag.Errorf("error fetching HCS Cluster (Managed Application ID %q) : %+v", managedAppID, err)
	}

	// Retrieve the valid upgrade versions
	upgradeVersionsResponse, err := meta.(*clients.Client).CustomResourceProvider.ListUpgradeVersions(ctx, *managedApp.ManagedResourceGroupID)
	if err != nil {
		return diag.Errorf("error retrieving upgrade versions for HCS Cluster (Managed Application ID %q): %+v", managedAppID, err)
	}

	newConsulVersion := d.Get("consul_version").(string)

	if upgradeVersionsResponse.Versions == nil {
		msg := "no upgrade versions of Consul are available for this cluster; you may already be on the latest Consul version supported by HCS"
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  msg,
				Detail:   msg,
			},
		}
	}

	if !consul.IsValidVersion(newConsulVersion, consul.FromAMAVersions(upgradeVersionsResponse.Versions)) {
		return diag.Errorf("specified Consul version (%s) is unavailable; must be one of: %+v", newConsulVersion, upgradeVersionsResponse.Versions)
	}

	updateResponse, err := meta.(*clients.Client).CustomResourceProvider.UpdateCluster(ctx, *managedApp.ManagedResourceGroupID, newConsulVersion)
	if err != nil {
		return diag.Errorf("error updating HCS Cluster (Managed Application ID %q) (Consul Version %s): %+v", managedAppID, newConsulVersion, err)
	}

	// TODO: Poll operation once that func lands in main
	log.Print(updateResponse.Operation.ID)

	return resourceClusterRead(ctx, d, meta)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Delete the managed app (the cluster custom resource will be deleted as well).
	managedAppID := d.Id()
	managedAppClient := meta.(*clients.Client).ManagedApplication
	future, err := managedAppClient.DeleteByID(ctx, managedAppID)
	if err != nil {
		return diag.Errorf("failed to delete HCS Cluster (Managed Application ID %q): %+v", managedAppID, err)
	}

	err = future.WaitForCompletionRef(ctx, managedAppClient.Client)
	if err != nil {
		return diag.Errorf("failed to wait for deleting HCS Cluster (Managed Application ID %q): %+v", managedAppID, err)
	}

	return nil
}

// setClusterResourceData sets the KV pairs of the cluster resource schema.
// We do not set consul_root_token_accessor_id and consul_root_token_secret_id here since
// the original root token is only available during cluster creation.
func setClusterResourceData(d *schema.ResourceData, managedApp managedapplications.Application, cluster models.HashicorpCloudConsulamaAmaClusterResponse) diag.Diagnostics {
	resourceGroupName, err := helper.ParseResourceGroupNameFromID(*managedApp.ID)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("resource_group_name", resourceGroupName)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("managed_application_name", *managedApp.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("email", cluster.Properties.Email)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set cluster mode based on numServers
	// TODO: cluster.Properties.ConsulClusterMode should be relied on when the value is populated on the fetch response
	clusterMode := "PRODUCTION"
	if cluster.Properties.ConsulNumServers == "1" {
		clusterMode = "DEVELOPMENT"
	}

	err = d.Set("cluster_mode", clusterMode)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("cluster_name", cluster.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("vnet_cidr", cluster.Properties.ConsulVnetCidr)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("consul_version", cluster.Properties.ConsulCurrentVersion)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("consul_datacenter", cluster.Properties.ConsulDatacenter)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("consul_federation_token", cluster.Properties.FederationToken)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("consul_external_endpoint", strings.ToLower(cluster.Properties.ConsulExternalEndpoint) == "enabled")
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("location", cluster.Properties.Location)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("plan_name", *managedApp.Plan.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	managedResourceGroupName, err := helper.ParseResourceGroupNameFromID(*managedApp.ManagedResourceGroupID)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("managed_resource_group_name", managedResourceGroupName)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("state", cluster.Properties.State)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("storage_account_name", cluster.Properties.StorageAccountName)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("blob_container_name", cluster.Properties.BlobContainerName)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("managed_application_id", cluster.Properties.ManagedAppID)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("storage_account_resource_group", cluster.Properties.StorageAccountResourceGroup)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("consul_automatic_upgrades", strings.ToLower(cluster.Properties.ConsulAutomaticUpgrades) == "enabled")
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("consul_snapshot_interval", cluster.Properties.ConsulSnapshotInterval)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("consul_snapshot_retention", cluster.Properties.ConsulSnapshotRetention)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("consul_config_file", cluster.Properties.ConsulConfigFile)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("consul_ca_file", cluster.Properties.ConsulCaFile)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("consul_connect", strings.ToLower(cluster.Properties.ConsulConnect) == "enabled")
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("consul_external_endpoint_url", cluster.Properties.ConsulExternalEndpointURL)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("consul_private_endpoint_url", cluster.Properties.ConsulPrivateEndpointURL)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("consul_cluster_id", cluster.Properties.ConsulClusterID)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
