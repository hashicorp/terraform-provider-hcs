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
	"github.com/hashicorp/terraform-provider-hcs/internal/helper"
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
		Description: "The cluster root token resource is the token used to bootstrap the cluster's ACL system." +
			" Using this resource to create a new root token for an cluster resource will invalidate the consul root token accessor id and consul root token secret id properties of the cluster.",
		CreateContext: resourceClusterRootTokenCreate,
		ReadContext:   resourceClusterRootTokenRead,
		DeleteContext: resourceClusterRootTokenDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultClusterRootTokenTimeoutDuration,
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
			// Computed outputs
			"accessor_id": {
				Description: "The accessor ID of the root ACL token.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"secret_id": {
				Description: "The secret ID of the root ACL token.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			"kubernetes_secret": {
				Description: "The root ACL token Base64 encoded in a Kubernetes secret.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
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
		if helper.IsAutoRestResponseCodeNotFound(app.Response) {
			// No managed application exists, so we should not try to create a root token
			return diag.Errorf("unable to create root token; no HCS Cluster found for (Managed Application %q) (Resource Group %q) (Correlation ID %q)",
				managedAppName,
				resourceGroupName,
				meta.(*clients.Client).CorrelationRequestID,
			)
		}

		return diag.Errorf("error checking for presence of existing HCS Cluster (Managed Application %q) (Resource Group %q) (Correlation ID %q): %+v",
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
		return diag.Errorf("error creating HCS Cluster root token (Managed Application %q) (Resource Group %q) (Correlation ID %q): %+v",
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
	managedApp, err := managedAppClient.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		if helper.IsAutoRestResponseCodeNotFound(managedApp.Response) {
			// No managed application exists, so this root token should be removed from state
			log.Printf("[WARN] no HCS Cluster found for (Managed Application %q) (Resource Group %q) (Correlation ID %q); removing root token.",
				managedAppName,
				resourceGroupName,
				meta.(*clients.Client).CorrelationRequestID,
			)
			d.SetId("")
			return nil
		}

		return diag.Errorf("error checking for presence of existing HCS Cluster (Managed Application %q) (Resource Group %q) (Correlation ID %q): %+v",
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
		if helper.IsAutoRestResponseCodeNotFound(app.Response) {
			// No managed application exists, so this root token should be removed from state
			log.Printf("[WARN] no HCS Cluster found for (Managed Application %q) (Resource Group %q) (Correlation ID %q)",
				managedAppName,
				resourceGroupName,
				meta.(*clients.Client).CorrelationRequestID,
			)
			return nil
		}

		return diag.Errorf("error checking for presence of existing HCS Cluster (Managed Application %q) (Resource Group %q) (Correlation ID %q): %+v",
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
		return diag.Errorf("error deleting HCS Cluster root token (Managed Application %q) (Resource Group %q) (Correlation ID %q): %+v",
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
