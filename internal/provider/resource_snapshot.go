package provider

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcs/internal/clients"
	"github.com/hashicorp/terraform-provider-hcs/internal/clients/hcs-ama-api-spec/models"
	"github.com/hashicorp/terraform-provider-hcs/internal/timeouts"
)

const (
	// defaultRestoredAt is the default string returned when a snapshot has not been restored
	defaultRestoredAt = "0001-01-01T00:00:00.000Z"
)

// resourceSnapshot defines the snapshot resource schema and CRUD contexts.
func resourceSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSnapshotCreate,
		ReadContext:   resourceSnapshotRead,
		UpdateContext: resourceSnapshotUpdate,
		DeleteContext: resourceSnapshotDelete,
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
				ValidateDiagFunc: validateManagedAppName,
			},
			"snapshot_name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateStringNotEmpty,
			},
			// Computed outputs
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"requested_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"finished_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"restored_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctx, cancel := timeouts.ForCreateUpdate(ctx, d)
	defer cancel()

	resourceGroupName := d.Get("resource_group_name").(string)
	managedAppName := d.Get("managed_application_name").(string)

	managedAppClient := meta.(*clients.Client).ManagedApplication
	app, err := managedAppClient.Get(ctx, resourceGroupName, managedAppName)
	// TODO handle 404 not found
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to fetch managed app: %s", err))
	}

	managedAppManagedResourceGroupID := *app.ManagedResourceGroupID
	snapshotName := d.Get("snapshot_name").(string)

	crpClient := meta.(*clients.Client).CustomResourceProvider
	resp, err := crpClient.CreateSnapshot(ctx, managedAppManagedResourceGroupID,
		resourceGroupName, snapshotName)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create snapshot for managed app: %s", err))
	}

	d.SetId(resp.SnapshotID)

	err = crpClient.PollOperation(ctx, resp.Operation.ID, managedAppManagedResourceGroupID, managedAppName, 10)

	if err != nil {
		log.Printf("[ERROR] - error polling operation!")
		return []diag.Diagnostic{
			{
				Severity:      0,
				Summary:       err.Error(),
				Detail:        resp.Operation.ID,
				AttributePath: nil,
			},
		}
	}

	return resourceSnapshotRead(ctx, d, meta)
}

func resourceSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctx, cancel := timeouts.ForRead(ctx, d)
	defer cancel()

	resourceGroupName := d.Get("resource_group_name").(string)
	managedAppName := d.Get("managed_application_name").(string)

	managedAppClient := meta.(*clients.Client).ManagedApplication
	app, err := managedAppClient.Get(ctx, resourceGroupName, managedAppName)
	// TODO handle 404 not found
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to fetch managed app: %s", err))
	}

	managedAppManagedResourceGroupID := *app.ManagedResourceGroupID
	snapshotID := d.Id()

	crpClient := meta.(*clients.Client).CustomResourceProvider
	resp, err := crpClient.GetSnapshot(ctx, managedAppManagedResourceGroupID,
		resourceGroupName, snapshotID)

	// TODO check if we get a 404 here and if so set ID to "" to delete from tfstate
	// TODO present error message to user to remove from their tf file
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get snapshot for managed app: %s", err))
	}

	populateSnapshotState(d, resp.Snapshot)

	return nil
}

func resourceSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctx, cancel := timeouts.ForUpdate(ctx, d)
	defer cancel()

	resourceGroupName := d.Get("resource_group_name").(string)
	managedAppName := d.Get("managed_application_name").(string)

	managedAppClient := meta.(*clients.Client).ManagedApplication
	app, err := managedAppClient.Get(ctx, resourceGroupName, managedAppName)
	// TODO handle 404 not found
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to fetch managed app: %s", err))
	}

	managedResourceGroupID := *app.ManagedResourceGroupID
	snapshotName := d.Get("snapshot_name").(string)
	snapshotID := d.Id()

	crpClient := meta.(*clients.Client).CustomResourceProvider
	resp, err := crpClient.RenameSnapshot(ctx, managedResourceGroupID, resourceGroupName, snapshotID, snapshotName)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to rename snapshot for managed app: %s", err))
	}

	populateSnapshotState(d, resp.Snapshot)

	return nil
}

func resourceSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("not implemented")
}

func populateSnapshotState(d *schema.ResourceData, snapshot *models.HashicorpCloudConsulamaAmaSnapshotProperties) {
	d.Set("state", snapshot.State)
	d.Set("requested_at", snapshot.RequestedAt.String())
	d.Set("finished_at", snapshot.FinishedAt.String())

	var size = 0
	size, err := strconv.Atoi(snapshot.Size)
	if err != nil {
		log.Printf("[ERROR] Error converting string to int: %v", err)
	}
	d.Set("size", size)

	if snapshot.RestoredAt.String() != defaultRestoredAt {
		d.Set("restored_at", snapshot.RestoredAt.String())
	}
}
