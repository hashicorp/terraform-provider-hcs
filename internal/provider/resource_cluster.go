package provider

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-05-01/network"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-07-01/managedapplications"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcs/internal/clients"
	"github.com/hashicorp/terraform-provider-hcs/internal/clients/hcs-ama-api-spec/models"
	"github.com/hashicorp/terraform-provider-hcs/internal/consul"
	"github.com/hashicorp/terraform-provider-hcs/internal/hcsmeta"
	"github.com/hashicorp/terraform-provider-hcs/internal/helper"
)

// createUpdateTimeoutDuration is the amount of time that can elapse
// before a cluster create or update operation should timeout.
var createUpdateTimeoutDuration = time.Minute * 60

// deleteTimeoutDuration is the amount of time that can elapse
// before a cluster delete operation should timeout.
var deleteTimeoutDuration = time.Minute * 25

// managedAppParamValue is the container struct for passing AMA values on creation/update.
type managedAppParamValue struct {
	// Value is the value of the AMA param
	Value interface{} `json:"value"`
}

// resourceCluster represents an HCS Cluster resource.
// Most of the CRUD involves the Azure Managed Application and Custom Resource Provider actions.
func resourceCluster() *schema.Resource {
	return &schema.Resource{
		Description:   "The cluster resource allows you to manage an HCS Azure Managed Application.",
		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		UpdateContext: resourceClusterUpdate,
		DeleteContext: resourceClusterDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultClusterTimeoutDuration,
			Create:  &createUpdateTimeoutDuration,
			Update:  &createUpdateTimeoutDuration,
			Delete:  &deleteTimeoutDuration,
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceClusterImport,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"resource_group_name": {
				Description:      "The name of the Resource Group in which the HCS Azure Managed Application belongs.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateResourceGroupName,
			},
			"managed_application_name": {
				Description:      "The name of the HCS Azure Managed Application.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"email": {
				Description:      "The contact email for the primary owner of the cluster.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validateStringNotEmpty,
			},
			"cluster_mode": {
				Description: "The mode of the cluster ('Development' or 'Production'). Development clusters only have a single Consul server. Production clusters are fully supported, full featured, and deploy with a minimum of three hosts.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
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
				Description:      "The name of the cluster Managed Resource. If not specified, it is defaulted to the value of `managed_application_name`.",
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Computed:         true,
				ValidateDiagFunc: validateSlugID,
			},
			"vnet_cidr": {
				Description:      "The VNET CIDR range of the Consul cluster.",
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          "172.25.16.0/24",
				ValidateDiagFunc: validateCIDR,
			},
			"min_consul_version": {
				Description:      "The minimum Consul version of the cluster. If not specified, it is defaulted to the version that is currently recommended by HCS.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateSemVer,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					// Suppress diff is normalized versions match OR min_consul_version is removed from the resource
					// since min_consul_version is required in order to upgrade the cluster to a new Consul version.
					return consul.NormalizeVersion(old) == consul.NormalizeVersion(new) || new == ""
				},
			},
			"consul_datacenter": {
				Description:      "The Consul data center name of the cluster. If not specified, it is defaulted to the value of `managed_application_name`.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateSlugID,
				ForceNew:         true,
				Computed:         true,
			},
			"consul_federation_token": {
				Description: "The token used to join a federation of Consul clusters. If the cluster is not part of a federation, this field will be empty.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					// Since federation tokens are not persisted in HCS, we generate a new one for each federation
					// token data source read. We don't want to force recreation of the cluster if the 'Primary' claim
					// of the 'new' JWT (federation token) matches the 'Primary' claim for the old token.
					return helper.FederationTokensHaveSamePrimary(old, new)
				},
			},
			"consul_external_endpoint": {
				Description: "Denotes that the cluster has an external endpoint for the Consul UI.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
			},
			"location": {
				Description: "The Azure region that the cluster is deployed to. If not specified, it is defaulted to the region of the Resource Group the Managed Application belongs to.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				// TODO: validate location the same way azurerm does
			},
			"plan_name": {
				Description: "The name of the Azure Marketplace HCS plan for the cluster. If not specified, it will default to the current HCS default plan (see the `hcs_plan_defaults` data source).",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				ValidateDiagFunc: validateStringInSlice([]string{
					"on-demand",
					"on-demand-v2",
					"annual",
				}, false),
			},
			"managed_resource_group_name": {
				Description: "The name of the Managed Resource Group in which the cluster resources belong. If not specified, it is defaulted to the value of `managed_application_name` with 'mrg-' prepended.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
			},
			"tags": {
				Description:      "A mapping of tags to assign to the HCS Azure Managed Application resource.",
				Type:             schema.TypeMap,
				Optional:         true,
				ValidateDiagFunc: validateAzureTags,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"audit_logging_enabled": {
				Description: "Enables Consul audit logging for the cluster resource.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"audit_log_storage_container_url": {
				Description: "The url of the Azure blob storage container to write audit logs to if `audit_logging_enabled` is `true`.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"managed_identity_name": {
				Description: "The name of the managed identity used for writing audit logs if `audit_logging_enable` is `true`.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			// Computed outputs
			"vnet_id": {
				Description: "The ID of the cluster's managed VNet.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vnet_name": {
				Description: "The name of the cluster's managed VNet.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vnet_resource_group_name": {
				Description: "The resource group that the cluster's managed VNet belongs to. This will be the same value as `managed_resource_group_name`.",
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
			"consul_version": {
				Description: "The Consul version of the cluster.",
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
			"consul_root_token_accessor_id": {
				Description: "The accessor ID of the root ACL token that is generated upon cluster creation. If a new root token is generated using the `hcs_cluster_root_token` resource, this field is no longer valid.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"consul_root_token_secret_id": {
				Description: "The secret ID of the root ACL token that is generated upon cluster creation. If a new root token is generated using the `hcs_cluster_root_token` resource, this field is no longer valid.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
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
		if !helper.IsAutoRestResponseCodeNotFound(existingCluster.Response) {
			return diag.Errorf("unable to check for presence of existing HCS cluster (Managed Application %q) (Resource Group %q) (Correlation ID %q): %v",
				managedAppName,
				resourceGroupName,
				meta.(*clients.Client).CorrelationRequestID,
				err,
			)
		}
	}
	if existingCluster.ID != nil && *existingCluster.ID != "" {
		return diag.Errorf("unable to create HCS cluster (%s) - an HCS cluster with this ID already exists; see resouce documentation for hcs_cluster for instructions on how to add an already existing HCS cluster to the state", *existingCluster.ID)
	}

	// Fetch resource group
	resourceGroup, err := meta.(*clients.Client).ResourceGroup.Get(ctx, resourceGroupName)
	if err != nil {
		return diag.Errorf("unable to fetch resource group (Resource Group %q) (Correlation ID %q): %v",
			resourceGroupName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	location := resourceGroup.Location
	v, ok := d.GetOk("location")
	if ok {
		location = helper.String(strings.ReplaceAll(strings.ToLower(v.(string)), " ", ""))
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
		return diag.Errorf("unable to fetch available HCP Consul versions: %v", err)
	}
	consulVersion := consul.RecommendedVersion(availableConsulVersions)
	v, ok = d.GetOk("min_consul_version")
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
		Name:      helper.String(planName),
		Version:   helper.String(planDefaults.Version),
		Product:   helper.String(meta.(*clients.Client).Config.MarketPlaceProductName),
		Publisher: helper.String("hashicorp-4665790"),
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

	var auditLogBucketURL string
	v, ok = d.GetOk("audit_log_storage_container_url")
	if ok {
		auditLogBucketURL = v.(string)
	}

	auditLoggingEnabled := "disabled"
	if d.Get("audit_logging_enabled").(bool) {
		if auditLogBucketURL == "" {
			return diag.Errorf("audit_log_storage_container_url must be set when audit_logging_enabled is true")
		}

		auditLoggingEnabled = "enabled"
	}

	hcsAMAParams := map[string]managedAppParamValue{
		"clusterMode": {
			Value: strings.ToUpper(d.Get("cluster_mode").(string)),
		},
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
		"sourceChannel": {
			Value: meta.(*clients.Client).Config.SourceChannel,
		},
		"auditLoggingEnabled": {
			Value: auditLoggingEnabled,
		},
		"auditLogStorageContainerURL": {
			Value: auditLogBucketURL,
		},
	}

	if federationToken != "" {
		hcsAMAParams["federationToken"] = managedAppParamValue{
			Value: federationToken,
		}
	}

	var tags map[string]*string
	v, ok = d.GetOk("tags")
	if ok {
		t := v.(map[string]interface{})
		tags = make(map[string]*string, len(t))

		for i, v := range t {
			tag, _ := helper.TagValueToString(v)
			tags[i] = &tag
		}
	}

	params := managedapplications.Application{
		ApplicationProperties: &managedapplications.ApplicationProperties{
			ManagedResourceGroupID: helper.String(managedResourceGroupId),
			Parameters:             hcsAMAParams,
		},
		Plan:     &plan,
		Kind:     helper.String("MarketPlace"),
		Location: location,
		Tags:     tags,
	}
	future, err := managedAppClient.CreateOrUpdate(ctx, resourceGroupName, managedAppName, params)
	if err != nil {
		return diag.Errorf("unable to create HCS cluster (Managed Application %q) (Resource Group %q) (Correlation ID %q): %v",
			managedAppName,
			resourceGroupName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}
	if err = future.WaitForCompletionRef(ctx, managedAppClient.Client); err != nil {
		return diag.Errorf("unable to wait for creation of HCS cluster (Managed Application %q) (Resource Group %q) (Correlation ID %q): %v",
			managedAppName,
			resourceGroupName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	app, err := managedAppClient.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		return diag.Errorf("unable to retrieve HCS cluster (Managed Application %q) (Resource Group %q) (Correlation ID %q): %v",
			managedAppName,
			resourceGroupName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}
	if app.ID == nil || *app.ID == "" {
		return diag.Errorf("unable to read HCS cluster ID (Managed Application %q) (Resource Group %q)", managedAppName, resourceGroupName)
	}

	// We need to set the cluster name to be able to fetch the cluster on read
	err = d.Set("cluster_name", clusterName)

	d.SetId(*app.ID)

	rootTokenResp, err := meta.(*clients.Client).CustomResourceProvider.CreateRootToken(ctx, *app.ApplicationProperties.ManagedResourceGroupID)
	if err != nil {
		return diag.Errorf("unable to create HCS cluster root token (Managed Application %q) (Resource Group %q) (Correlation ID %q): %v",
			managedAppName,
			resourceGroupName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
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
		if helper.IsAutoRestResponseCodeNotFound(managedApp.Response) {
			log.Printf("[WARN] no HCS Cluster found for (Managed Application ID %q) (Correlation ID %q); removing from state",
				managedAppID,
				meta.(*clients.Client).CorrelationRequestID,
			)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to fetch HCS cluster (Managed Application ID %q) (Correlation ID %q): %v",
			managedAppID,
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
		return diag.Errorf("unable to fetch HCS cluster (Managed Application ID %q) (Cluster Name %q) (Correlation ID %q): %v",
			managedAppID,
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
		return diag.Errorf("unable to fetch VNet for HCS cluster (Managed Application ID %q) (Managed Resource Group Name %q) (VNet Name %q) (Correlation ID %q): %v",
			managedAppID,
			managedResourceGroupName,
			vNetName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	return setClusterData(d, managedApp, cluster, vNet)
}

func toModelBoolean(b bool) models.HashicorpCloudConsulamaAmaBoolean {
	if b {
		return models.HashicorpCloudConsulamaAmaBooleanTRUE
	}
	return models.HashicorpCloudConsulamaAmaBooleanFALSE
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Fetch the managed app
	managedAppID := d.Id()
	managedApp, err := meta.(*clients.Client).ManagedApplication.GetByID(ctx, managedAppID)
	if err != nil {
		if helper.IsAutoRestResponseCodeNotFound(managedApp.Response) {
			log.Printf("[WARN] no HCS Cluster found for (Managed Application ID %q) (Correlation ID %q); removing from state",
				managedAppID,
				meta.(*clients.Client).CorrelationRequestID,
			)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to fetch HCS cluster (Managed Application ID %q) (Correlation ID %q): %v",
			managedAppID,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	update := &models.HashicorpCloudConsulamaAmaClusterUpdate{}

	auditLoggingChanged := d.HasChange("audit_logging_enabled") || d.HasChange("audit_log_storage_container_url")

	if auditLoggingChanged {
		url := d.Get("audit_log_storage_container_url").(string)
		auditLoggingEnabled := d.Get("audit_logging_enabled").(bool)
		if auditLoggingEnabled && url == "" {
			return diag.Errorf("audit_log_storage_container_url must be set when audit_logging_enabled is true")
		}

		update.AuditLogging = &models.HashicorpCloudConsulamaAmaAuditLoggingUpdate{
			Enabled:             toModelBoolean(auditLoggingEnabled),
			StorageContainerURL: url,
		}
	}

	// If the min_consul_version differs from the current version, attempt to upgrade the cluster
	versionChanged := d.HasChange("min_consul_version")
	if versionChanged {
		update.ConsulVersion = d.Get("min_consul_version").(string)
	}

	// Only execute the UpdateCluster custom action on the managed app if the audit logging
	// configuration or the Consul version has been changed.
	if versionChanged || auditLoggingChanged {
		if err := upgradeCluster(ctx, meta, managedApp, update); err != nil {
			return err
		}
	}

	// If we are updating due to modified tags OR removing existing tags, attempt to update the Managed App
	_, ok := d.GetOk("tags")
	if ok || (!ok && len(managedApp.Tags) > 0) {
		managedAppUpdateDiag := updateManagedApplicationTags(ctx, d, meta, managedApp)
		if managedAppUpdateDiag != nil {
			return managedAppUpdateDiag
		}
	}

	return resourceClusterRead(ctx, d, meta)
}

// upgradeClusterVersion updates a cluster's Consul version to a valid upgrade version
func upgradeCluster(ctx context.Context, meta interface{}, managedApp managedapplications.Application, update *models.HashicorpCloudConsulamaAmaClusterUpdate) diag.Diagnostics {
	if update.ConsulVersion != "" {
		// Retrieve the valid upgrade versions
		upgradeVersionsResponse, err := meta.(*clients.Client).CustomResourceProvider.ListUpgradeVersions(ctx, *managedApp.ManagedResourceGroupID)
		if err != nil {
			return diag.Errorf("unable to retrieve upgrade versions for HCS cluster (Managed Application ID %q) (Correlation ID %q): %v",
				*managedApp.ID,
				meta.(*clients.Client).CorrelationRequestID,
				err,
			)
		}

		newConsulVersion := consul.NormalizeVersion(update.ConsulVersion)

		if upgradeVersionsResponse.Versions == nil {
			return diag.Errorf("no upgrade versions of Consul are available for this cluster; you may already be on the latest Consul version supported by HCS")
		}

		if !consul.IsValidVersion(newConsulVersion, consul.FromAMAVersions(upgradeVersionsResponse.Versions)) {
			return diag.Errorf("specified Consul version (%s) is unavailable; must be one of: %+v", newConsulVersion, consul.FromAMAVersions(upgradeVersionsResponse.Versions))
		}
	}

	updateResponse, err := meta.(*clients.Client).CustomResourceProvider.UpdateCluster(ctx, *managedApp.ManagedResourceGroupID, update)
	if err != nil {
		return diag.Errorf("unable to update HCS cluster (Managed Application ID %q) (Consul Version %s) (Correlation ID %q): %v",
			*managedApp.ID,
			update.ConsulVersion,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	err = meta.(*clients.Client).CustomResourceProvider.PollOperation(ctx, updateResponse.Operation.ID, *managedApp.ManagedResourceGroupID, *managedApp.Name, 10)
	if err != nil {
		return diag.Errorf("unable to poll update cluster operation (Managed Application ID %q) (Consul Version %s) (Correlation ID %q): %v",
			*managedApp.ID,
			update.ConsulVersion,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	return nil
}

// updateManagedApplicationTags updates a cluster's Managed Application tags
func updateManagedApplicationTags(ctx context.Context, d *schema.ResourceData, meta interface{}, managedApp managedapplications.Application) diag.Diagnostics {
	t := d.Get("tags").(map[string]interface{})
	tags := make(map[string]*string, len(t))

	for i, v := range t {
		tag, _ := helper.TagValueToString(v)
		tags[i] = &tag
	}

	updateResp, err := meta.(*clients.Client).ManagedApplication.Update(
		ctx,
		d.Get("resource_group_name").(string),
		*managedApp.Name,
		&managedapplications.ApplicationPatchable{Tags: tags},
	)
	if err != nil {
		// Azure seems to return a 202 on successful update, but the Autorest client has trouble responding
		// to the response. Ignore the error in this case as the update was successful.
		if helper.IsAutoRestResponseCodeAccepted(updateResp.Response) {
			log.Printf("[INFO] successfully updated tags for (Managed Application ID %q)", *managedApp.ID)
			return nil
		}

		return diag.Errorf("unable to update Managed Application tags (Managed Application ID %q) (Correlation ID %q): %v",
			*managedApp.ID,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	return nil
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Ensure the cluster is not the primary in a federation that still has secondaries
	managedAppID := d.Id()
	managedAppClient := meta.(*clients.Client).ManagedApplication
	managedApp, err := managedAppClient.GetByID(ctx, managedAppID)
	if err != nil {
		if helper.IsAutoRestResponseCodeNotFound(managedApp.Response) {
			log.Printf("[WARN] no HCS Cluster found for (Managed Application ID %q) (Correlation ID %q)",
				managedAppID,
				meta.(*clients.Client).CorrelationRequestID,
			)
			return nil
		}

		return diag.Errorf("unable to fetch HCS cluster before deletion (Managed Application ID %q) (Correlation ID %q): %v",
			managedAppID,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	resourceGroupName := d.Get("resource_group_name").(string)

	// An error here denotes that the cluster is not part of a federation
	federationResponse, err := meta.(*clients.Client).CustomResourceProvider.GetFederation(ctx, *managedApp.ManagedResourceGroupID, d.Get("resource_group_name").(string))
	// Ensure the cluster is not the primary in the federation
	if err == nil && isClusterPrimaryInFederation(*managedApp.Name, resourceGroupName, federationResponse) {
		return diag.Errorf("unable to delete primary datacenter of a federation before all secondary datacenters are deleted: (Managed Application %q) (Resource Group %q)", *managedApp.Name, resourceGroupName)
	}

	// Delete the managed app (the cluster custom resource will be deleted as well).
	future, err := managedAppClient.DeleteByID(ctx, managedAppID)
	if err != nil {
		return diag.Errorf("unable to delete HCS cluster (Managed Application ID %q) (Correlation ID %q): %v",
			managedAppID,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	err = future.WaitForCompletionRef(ctx, managedAppClient.Client)
	if err != nil {
		return diag.Errorf("unable to wait for delete of HCS cluster (Managed Application ID %q) (Correlation ID %q): %v",
			managedAppID,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	// Sleep to prevent ResourcePurchaseCanceling errors returned from Azure during the scenario when
	// a cluster resource must be deleted and re-created.
	time.Sleep(time.Minute)

	return nil
}

// isClusterPrimaryInFederation determines if a cluster's managed app and resource group names match
// the primary cluster's managed app and resource group names in a non-empty federation.
func isClusterPrimaryInFederation(managedAppName string, resourceGroupName string, federationResponse models.HashicorpCloudConsulamaAmaGetFederationResponse) bool {
	if federationResponse.PrimaryDatacenter == nil || len(federationResponse.SecondaryDatacenters) == 0 {
		return false
	}

	return federationResponse.PrimaryDatacenter.Name == managedAppName && federationResponse.PrimaryDatacenter.ResourceGroup == resourceGroupName
}

// resourceClusterImport implements the logic necessary to import an un-tracked
// (by Terraform) cluster resource into Terraform state.
//
// This logic handles parsing out the AMA ID + cluster name to build the proper
// request to fetch the cluster details.
func resourceClusterImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	id, clusterName, err := validateClusterImportString(d.Id())

	if err != nil {
		return nil, err
	}

	d.SetId(id)
	d.Set("cluster_name", clusterName)

	diags := resourceClusterRead(ctx, d, meta)
	if err := helper.ToError(diags); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

// validateClusterImportString validates that the import string
// is of the format expected.  Which should be a colon `:` delimited
// string with the managed_application_id to the left of the colon
// and the cluster_name to the right of it:
//
// `managed_application_id:cluster_name`
func validateClusterImportString(s string) (string, string, error) {
	if !strings.Contains(s, ":") {
		return "", "", fmt.Errorf("import id string must be of format `managed_application_id:cluster_name`; id string: %s does not contain `:`", s)
	}

	segments := strings.Split(s, ":")
	if len(segments) != 2 {
		return "", "", fmt.Errorf("import id string must be of format `managed_application_id:cluster_name`; id string: %s contains more than one `:`", s)
	}

	if segments[0] == "" {
		return "", "", fmt.Errorf("import id string must be of format `managed_application_id:cluster_name`; id string: %s has empty string to left of `:`", s)
	}

	if segments[1] == "" {
		return "", "", fmt.Errorf("import id string must be of format `managed_application_id:cluster_name`; id string: %s has empty string to right of `:`", s)
	}

	return segments[0], segments[1], nil
}

// setClusterData sets the KV pairs of the cluster resource schema.
// We do not set consul_root_token_accessor_id and consul_root_token_secret_id here since
// the original root token is only available during cluster creation.
func setClusterData(d *schema.ResourceData, managedApp managedapplications.Application, cluster models.HashicorpCloudConsulamaAmaClusterResponse, vNet network.VirtualNetwork) diag.Diagnostics {
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

	err = d.Set("vnet_id", *vNet.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("vnet_name", *vNet.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	managedResourceGroupName, err := helper.ParseResourceGroupNameFromID(*managedApp.ManagedResourceGroupID)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("vnet_resource_group_name", managedResourceGroupName)
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

	err = d.Set("managed_resource_group_name", managedResourceGroupName)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("tags", helper.FlattenTags(managedApp.Tags))
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

	auditLoggingEnabled := cluster.Properties.AuditLoggingEnabled == models.HashicorpCloudConsulamaAmaBooleanTRUE
	err = d.Set("audit_logging_enabled", auditLoggingEnabled)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("audit_log_storage_container_url", cluster.Properties.AuditLogStorageContainerURL)
	if err != nil {
		return diag.FromErr(err)
	}

	managedIdentityName := helper.ParseResourceNameFromID(cluster.Properties.ManagedIdentity)
	err = d.Set("managed_identity_name", managedIdentityName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
