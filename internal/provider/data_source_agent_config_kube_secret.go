package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcs/internal/clients"
)

// agentConfigKubernetesSecretTemplate is the template used to generate a
// Kubernetes formatted secret for the Consul agent config.
const agentConfigKubernetesSecretTemplate = `apiVersion: v1
kind: Secret
metadata:
  name: %s-hcs
type: Opaque
data:
  gossipEncryptionKey: %s
  caCert: %s`

// defaultAgentConfigKubernetesSecretTimeoutDuration is the default timeout
// for reading the agent config Kubernetes secret.
var defaultAgentConfigKubernetesSecretTimeoutDuration = time.Minute * 5

// consulConfig represents the Consul config returned on the GetConfig response.
type consulConfig struct {
	GossipKey string `json:"encrypt"`
}

// dataSourceAgentConfigKubernetesSecret is the data source for the Consul versions supported by HCS.
func dataSourceAgentConfigKubernetesSecret() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAgentConfigKubernetesSecretRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultAgentConfigKubernetesSecretTimeoutDuration,
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
			// Computed output
			"secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

// dataSourceAgentConfigKubernetesSecretRead retrieves the Consul config and formats a Kubernetes secret for Consul agents running
// in Kubernetes to leverage.
func dataSourceAgentConfigKubernetesSecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	managedAppName := d.Get("managed_application_name").(string)
	resourceGroupName := d.Get("resource_group_name").(string)

	managedApp, err := meta.(*clients.Client).ManagedApplication.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		if managedApp.Response.StatusCode == 404 {
			return diag.Errorf("HCS Cluster (Resource Group Name %q) (Managed Application Name %q) was not found", resourceGroupName, managedAppName)
		}

		return diag.Errorf("error fetching HCS Cluster (Resource Group Name %q) (Managed Application Name %q) : %+v", resourceGroupName, managedAppName, err)
	}

	config, caFile, err := meta.(*clients.Client).CustomResourceProvider.GetConsulConfig(ctx, *managedApp.ManagedResourceGroupID, resourceGroupName)
	if err != nil {
		return diag.Errorf("error fetching Consul config (Resource Group Name %q) (Managed Application Name %q) : %+v", resourceGroupName, managedAppName, err)
	}

	encodedGossipKey := base64.StdEncoding.EncodeToString([]byte(config.GossipKey))

	encodedCAFile := base64.StdEncoding.EncodeToString([]byte(caFile))

	err = d.Set("secret", fmt.Sprintf(agentConfigKubernetesSecretTemplate, managedAppName, encodedGossipKey, encodedCAFile))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*managedApp.ID)

	return nil
}
