package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcs/internal/clients"
)

// helmConfigTemplate is the template used to generate a helm
// config for an AKS cluster based on given inputs.
//
// see generateHelmConfig for details on the inputs passed in
const helmConfigTemplate = `global:
  enabled: false
  name: consul
  datacenter: %s
  acls:
    manageSystemACLs: true
    bootstrapToken:
      secretName: %s-bootstrap-token
      secretKey: token
  gossipEncryption:
    secretName: %s-hcs
    secretKey: gossipEncryptionKey
  tls:
    enabled: true
    enableAutoEncrypt: true
    caCert:
      secretName: %s-hcs
      secretKey: caCert
externalServers:
  enabled: true
  hosts: %s
  httpsPort: 443
  useSystemRoots: true
  k8sAuthMethodHost: https://%s:443
client:
  enabled: true
  exposeGossipPorts: %t
  join: %s
connectInject:
  enabled: true`

// dataSourceAgentHelmConfig is the data source for the agent Helm
// config for an HCS cluster.
func dataSourceAgentHelmConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAgentHelmConfigRead,
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
			"aks_cluster_name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateStringNotEmpty,
			},
			// Optional
			"aks_resource_group": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateStringNotEmpty,
			},
			"expose_gossip_ports": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			// Computed outputs
			"config": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// dataSourceAgentHelmConfigRead is the func to implement reading of the
// agent Helm config for an HCS cluster.
func dataSourceAgentHelmConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resourceGroupName := d.Get("resource_group_name").(string)
	managedAppName := d.Get("managed_application_name").(string)

	managedAppClient := meta.(*clients.Client).ManagedApplication
	app, err := managedAppClient.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		if app.Response.StatusCode == 404 {
			// No managed application exists, so returning an error stating as such
			return diag.Errorf("no HCS Cluster found for (Managed Application %q) (Resource Group %q).", managedAppName, resourceGroupName)
		}

		return diag.Errorf("failed to check for presence of existing HCS Cluster (Managed Application %q) (Resource Group %q): %+v", managedAppName, resourceGroupName, err)
	}

	managedAppManagedResourceGroupID := *app.ManagedResourceGroupID

	crpClient := meta.(*clients.Client).CustomResourceProvider

	consulConfig, _, err := crpClient.GetConsulConfig(ctx, managedAppManagedResourceGroupID, resourceGroupName)
	if err != nil {
		return diag.Errorf("failed to get config for managed app: %+v", err)
	}

	// default to resource group name if aks_resource_group not present
	aksResourceGroup := resourceGroupName
	v, ok := d.GetOk("aks_resource_group")
	if ok {
		aksResourceGroup = v.(string)
	}

	aksClusterName := d.Get("aks_cluster_name").(string)

	mcClient := meta.(*clients.Client).ManagedClusters

	mcResp, err := mcClient.Get(ctx, aksResourceGroup, aksClusterName)
	if err != nil {
		return diag.Errorf("failed to check for presence of existing AKS Cluster (Cluster name %q) (Resource Group %q): %+v", aksClusterName, aksResourceGroup, err)
	}
	if mcResp.Response.StatusCode == 404 {
		// No AKS cluster exists, so returning an error stating as such
		return diag.Errorf("no AKS Cluster found for (Cluster name %q) (Resource Group %q).", aksClusterName, aksResourceGroup)
	}

	var exposeGossipPorts bool
	v, ok = d.GetOk("expose_gossip_ports")
	if ok {
		exposeGossipPorts = v.(bool)
	}

	if err := d.Set("config", generateHelmConfig(
		managedAppName, consulConfig.Datacenter, *mcResp.Fqdn, consulConfig.RetryJoin, exposeGossipPorts)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*app.ID + "/agent-helm-config")

	return nil
}

// generateHelmConfig will generate a helm config based on the passed in
// name, data center, retry join, and fqdn.
func generateHelmConfig(name, datacenter, fqdn string, retryJoin []string, exposeGossipPorts bool) string {
	// lowercase the name
	lower := strings.ToLower(name)

	// print retryJoin a double-quoted string safely escaped with Go syntax
	rj := fmt.Sprintf("%q", retryJoin)

	// replace any escaped double-quotes with single quotes
	// this is to match the format the the HCS CLI is outputting
	rj = strings.Replace(rj, "\"", "'", -1)

	return fmt.Sprintf(helmConfigTemplate,
		datacenter,
		lower, lower, lower,
		rj,
		fqdn,
		exposeGossipPorts,
		rj,
	)
}
