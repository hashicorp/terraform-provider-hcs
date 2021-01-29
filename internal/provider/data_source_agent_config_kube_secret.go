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

// dataSourceAgentConfigKubernetesSecret is the data source for generating the configuration for a
// Consul agent in the form of a Kubernetes secret.
func dataSourceAgentConfigKubernetesSecret() *schema.Resource {
	return &schema.Resource{
		Description: "The agent config Kubernetes secret data source provides Consul agents running in Kubernetes the configuration needed to connect to the Consul cluster.",
		ReadContext: dataSourceAgentConfigKubernetesSecretRead,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultAgentConfigKubernetesSecretTimeoutDuration,
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
			// Computed output
			"secret": {
				Description: "The Consul agent configuration in the format of a Kubernetes secret (YAML).",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
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
		return diag.Errorf("unable to fetch HCS cluster (Resource Group Name %q) (Managed Application Name %q) (Correlation ID %q): %v",
			resourceGroupName,
			managedAppName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	config, err := meta.(*clients.Client).CustomResourceProvider.GetConsulConfig(ctx, *managedApp.ManagedResourceGroupID, resourceGroupName)
	if err != nil {
		return diag.Errorf("unable to fetch Consul config (Resource Group Name %q) (Managed Application Name %q) (Correlation ID %q): %v",
			resourceGroupName,
			managedAppName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	encodedGossipKey := base64.StdEncoding.EncodeToString([]byte(config.GossipKey))

	encodedCAFile := base64.StdEncoding.EncodeToString([]byte(config.CaFile))

	err = d.Set("secret", fmt.Sprintf(agentConfigKubernetesSecretTemplate, managedAppName, encodedGossipKey, encodedCAFile))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*managedApp.ID)

	return nil
}
