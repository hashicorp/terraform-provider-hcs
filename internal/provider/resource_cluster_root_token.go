package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcs/internal/clients"
)

var defaultClusterRootTokenTimeoutDuration = time.Minute * 5

// rootTokenKubernetesSecretTemplate is the template used to generate a
// kubernetes formatted secret for the cluster root token.
const rootTokenKubernetesSecretTemplate = `apiVersion: v1
kind: Secret
metadata:
  name: %s-bootstrap-token
type: Opaque
data:
  token: %s`

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
		Description: "hcs_cluster_root_token is the token used to bootstrap the cluster's ACL system",
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
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"kubernetes_secret": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "the root token base64 encoded in a Kubernetes secret",
			},
		},
	}
}

// resourceClusterRootTokenCreate will generate a new root token for the cluster
func resourceClusterRootTokenCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resourceGroupName := d.Get("resource_group_name").(string)
	managedAppName := d.Get("managed_application_name").(string)

	managedAppClient := meta.(*clients.Client).ManagedApplication
	app, err := managedAppClient.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		if app.Response.StatusCode == 404 {
			// No managed application exists, so we should not try to create a root token
			return diag.Errorf("unable to create root token; no HCS Cluster found for (Managed Application %q) (Resource Group %q) (Correlation ID %q)",
				managedAppName,
				resourceGroupName,
				meta.(*clients.Client).CorrelationRequestID,
			)
		}

		return diag.Errorf("failed to check for presence of existing HCS Cluster (Managed Application %q) (Resource Group %q) (Correlation ID %q): %+v",
			managedAppName,
			resourceGroupName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	mrgID := *app.ApplicationProperties.ManagedResourceGroupID

	crpClient := meta.(*clients.Client).CustomResourceProvider
	rootTokenResp, err := crpClient.CreateRootToken(ctx, mrgID)
	if err != nil {
		return diag.Errorf("failed to create HCS Cluster root token (Managed Application %q) (Resource Group %q) (Correlation ID %q): %+v",
			managedAppName,
			resourceGroupName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	err = d.Set("accessor_id", rootTokenResp.MasterToken.AccessorID)
	if err != nil {
		return diag.FromErr(err)
	}

	secretID := rootTokenResp.MasterToken.SecretID
	err = d.Set("secret_id", secretID)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("kubernetes_secret", generateKubernetesSecret(secretID, managedAppName))
	if err != nil {
		return diag.FromErr(err)
	}

	// set the id to the value of the accessor id
	d.SetId(rootTokenResp.MasterToken.AccessorID)

	return nil
}

// resourceClusterRootTokenRead will act as a no-op as the root token is not persisted in
// any way that it can be fetched and read
func resourceClusterRootTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resourceGroupName := d.Get("resource_group_name").(string)
	managedAppName := d.Get("managed_application_name").(string)

	managedAppClient := meta.(*clients.Client).ManagedApplication
	app, err := managedAppClient.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		if app.Response.StatusCode == 404 {
			// No managed application exists, so this snapshot should be removed from state
			log.Printf("[WARN] no HCS Cluster found for (Managed Application %q) (Resource Group %q) (Correlation ID %q); removing root token.",
				managedAppName,
				resourceGroupName,
				meta.(*clients.Client).CorrelationRequestID,
			)
			d.SetId("")
			return nil
		}

		return diag.Errorf("failed to check for presence of existing HCS Cluster (Managed Application %q) (Resource Group %q) (Correlation ID %q): %+v",
			managedAppName,
			resourceGroupName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	return nil
}

// resourceClusterRootTokenDelete will "delete" an existing token by creating a new one,
// that will not be returned, and invalidating the previous token for the cluster.
func resourceClusterRootTokenDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resourceGroupName := d.Get("resource_group_name").(string)
	managedAppName := d.Get("managed_application_name").(string)

	managedAppClient := meta.(*clients.Client).ManagedApplication
	app, err := managedAppClient.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		if app.Response.StatusCode == 404 {
			// No managed application exists, so this root token should be removed from state
			log.Printf("[WARN] no HCS Cluster found for (Managed Application %q) (Resource Group %q) (Correlation ID %q)",
				managedAppName,
				resourceGroupName,
				meta.(*clients.Client).CorrelationRequestID,
			)
			return nil
		}

		return diag.Errorf("failed to check for presence of existing HCS Cluster (Managed Application %q) (Resource Group %q) (Correlation ID %q): %+v",
			managedAppName,
			resourceGroupName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	mrgID := *app.ApplicationProperties.ManagedResourceGroupID

	crpClient := meta.(*clients.Client).CustomResourceProvider
	// generate a new token to invalidate the previous one, but discard the response
	_, err = crpClient.CreateRootToken(ctx, mrgID)
	if err != nil {
		return diag.Errorf("failed to delete HCS Cluster root token (Managed Application %q) (Resource Group %q) (Correlation ID %q): %+v",
			managedAppName,
			resourceGroupName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	return nil
}

// generateKubernetesSecret will generate a Kubernetes secret with
// a base64 encoded root token secret as it's token.
func generateKubernetesSecret(rootTokenSecretId, managedAppName string) string {
	return fmt.Sprintf(rootTokenKubernetesSecretTemplate,
		// lowercase the name
		strings.ToLower(managedAppName),
		// base64 encode the secret value
		base64.StdEncoding.EncodeToString([]byte(rootTokenSecretId)))
}
