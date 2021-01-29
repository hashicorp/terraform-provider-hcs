package provider

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcs/internal/clients"
	"github.com/hashicorp/terraform-provider-hcs/internal/clients/hcs-ama-api-spec/models"
	"github.com/hashicorp/terraform-provider-hcs/internal/helper"
)

const (
	// defaultRestoredAt is the default string returned when a snapshot has not been restored
	defaultRestoredAt = "0001-01-01T00:00:00.000Z"
)

// defaultSnapshotTimeoutDuration is the amount of time that can elapse
// before a snapshot read should timeout.
var defaultSnapshotTimeoutDuration = time.Minute * 5

// snapshotCreateUpdateDeleteTimeoutDuration is the amount of time that can elapse
// before a snapshot operation should timeout.
var snapshotCreateUpdateDeleteTimeoutDuration = time.Minute * 15

// resourceSnapshot defines the snapshot resource schema and CRUD contexts.
func resourceSnapshot() *schema.Resource {
	return &schema.Resource{
		Description: "The snapshot resource allows users to manage Consul snapshots of an HCS cluster." +
			" Snapshots currently have a retention policy of 30 days.",
		CreateContext: resourceSnapshotCreate,
		ReadContext:   resourceSnapshotRead,
		UpdateContext: resourceSnapshotUpdate,
		DeleteContext: resourceSnapshotDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: &defaultSnapshotTimeoutDuration,
			Create:  &snapshotCreateUpdateDeleteTimeoutDuration,
			Update:  &snapshotCreateUpdateDeleteTimeoutDuration,
			Delete:  &snapshotCreateUpdateDeleteTimeoutDuration,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"resource_group_name": {
				Description:      "The name of the Resource Group in which the HCS Azure Managed Application belongs.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateResourceGroupName,
				ForceNew:         true,
			},
			"managed_application_name": {
				Description:      "The name of the HCS Azure Managed Application.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateManagedAppName,
				ForceNew:         true,
			},
			"snapshot_name": {
				Description:      "The name of the snapshot.",
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateStringNotEmpty,
			},
			// Computed outputs
			"state": {
				Description: "The state of the snapshot.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"size": {
				Description: "The size of the snapshot in bytes.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"requested_at": {
				Description: "Timestamp of when the snapshot was requested.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"finished_at": {
				Description: "Timestamp of when the snapshot was finished.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"restored_at": {
				Description: "Timestamp of when the snapshot was restored. If the snapshot has not been restored, this field will be blank.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resourceGroupName := d.Get("resource_group_name").(string)
	managedAppName := d.Get("managed_application_name").(string)

	managedAppClient := meta.(*clients.Client).ManagedApplication
	app, err := managedAppClient.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		if helper.IsAutoRestResponseCodeNotFound(app.Response) {
			// No managed application exists, so we should not try to create the snapshot
			return diag.Errorf("unable to create snapshot; HCS cluster not found (Managed Application %q) (Resource Group %q) (Correlation ID %q)",
				managedAppName,
				resourceGroupName,
				meta.(*clients.Client).CorrelationRequestID,
			)
		}

		return diag.Errorf("unable to check for presence of an existing HCS cluster (Managed Application %q) (Resource Group %q) (Correlation ID %q): %v",
			managedAppName,
			resourceGroupName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	managedAppManagedResourceGroupID := *app.ManagedResourceGroupID
	snapshotName := d.Get("snapshot_name").(string)

	crpClient := meta.(*clients.Client).CustomResourceProvider
	resp, err := crpClient.CreateSnapshot(ctx, managedAppManagedResourceGroupID,
		resourceGroupName, snapshotName)
	if err != nil {
		return diag.Errorf("unable to create snapshot (Managed Application %q) (Resource Group %q) (Correlation ID %q): %v",
			managedAppName,
			resourceGroupName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	d.SetId(resp.SnapshotID)

	err = crpClient.PollOperation(ctx, resp.Operation.ID, managedAppManagedResourceGroupID, managedAppName, 10)

	if err != nil {
		return diag.Errorf("unable to poll create snapshot operation (Managed Application %q) (Resource Group %q) (Correlation ID %q): %v",
			managedAppName,
			resourceGroupName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	return resourceSnapshotRead(ctx, d, meta)
}

func resourceSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resourceGroupName := d.Get("resource_group_name").(string)
	managedAppName := d.Get("managed_application_name").(string)

	managedAppClient := meta.(*clients.Client).ManagedApplication
	app, err := managedAppClient.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		if helper.IsAutoRestResponseCodeNotFound(app.Response) {
			// No managed application exists, so this snapshot should be removed from state
			log.Printf("[WARN] no HCS Cluster found for (Managed Application %q) (Resource Group %q) (Correlation ID %q)",
				managedAppName,
				resourceGroupName,
				meta.(*clients.Client).CorrelationRequestID,
			)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to check for presence of an existing HCS cluster (Managed Application %q) (Resource Group %q) (Correlation ID %q): %v",
			managedAppName,
			resourceGroupName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	managedAppManagedResourceGroupID := *app.ManagedResourceGroupID
	snapshotID := d.Id()

	crpClient := meta.(*clients.Client).CustomResourceProvider
	resp, err := crpClient.GetSnapshot(ctx, managedAppManagedResourceGroupID,
		resourceGroupName, snapshotID)

	if err != nil {
		if crpClient.IsCRPErrorAzureNotFound(err) {
			log.Printf("[WARN] snapshot not found. the retention policy for snapshots is 30 days and " +
				"this snapshot may have been deleted, if you leave the snapshot resource " +
				"in your plan, a new snapshot will be created")
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to fetch snapshot (Managed Application %q) (Resource Group %q) (Correlation ID %q): %v",
			managedAppName,
			resourceGroupName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	if diagnostics := populateSnapshotState(d, resp.Snapshot); diagnostics != nil {
		return diagnostics
	}

	return nil
}

func resourceSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resourceGroupName := d.Get("resource_group_name").(string)
	managedAppName := d.Get("managed_application_name").(string)

	managedAppClient := meta.(*clients.Client).ManagedApplication
	app, err := managedAppClient.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		if helper.IsAutoRestResponseCodeNotFound(app.Response) {
			// No managed application exists, so this snapshot should be removed from state
			log.Printf("[WARN] no HCS Cluster found for (Managed Application %q) (Resource Group %q) (Correlation ID %q)",
				managedAppName,
				resourceGroupName,
				meta.(*clients.Client).CorrelationRequestID,
			)
			d.SetId("")
			return nil
		}

		return diag.Errorf("unable to check for presence of an existing HCS cluster (Managed Application %q) (Resource Group %q) (Correlation ID %q): %v",
			managedAppName,
			resourceGroupName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	managedResourceGroupID := *app.ManagedResourceGroupID
	snapshotName := d.Get("snapshot_name").(string)
	snapshotID := d.Id()

	crpClient := meta.(*clients.Client).CustomResourceProvider
	resp, err := crpClient.RenameSnapshot(ctx, managedResourceGroupID, resourceGroupName, snapshotID, snapshotName)
	if err != nil {
		return diag.Errorf("unable to rename snapshot (Managed Application %q) (Resource Group %q) (Correlation ID %q): %v",
			managedAppName,
			resourceGroupName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	if diagnostics := populateSnapshotState(d, resp.Snapshot); diagnostics != nil {
		return diagnostics
	}

	return nil
}

func resourceSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resourceGroupName := d.Get("resource_group_name").(string)
	managedAppName := d.Get("managed_application_name").(string)

	managedAppClient := meta.(*clients.Client).ManagedApplication
	app, err := managedAppClient.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		if helper.IsAutoRestResponseCodeNotFound(app.Response) {
			// No managed application exists, so this snapshot should be removed from state
			log.Printf("[WARN] no HCS Cluster found for (Managed Application %q) (Resource Group %q) (Correlation ID %q)",
				managedAppName,
				resourceGroupName,
				meta.(*clients.Client).CorrelationRequestID,
			)
			return nil
		}

		return diag.Errorf("unable to check for presence of an existing HCS cluster (Managed Application %q) (Resource Group %q) (Correlation ID %q): %v",
			managedAppName,
			resourceGroupName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	managedAppManagedResourceGroupID := *app.ManagedResourceGroupID
	snapshotID := d.Id()

	crpClient := meta.(*clients.Client).CustomResourceProvider
	resp, err := crpClient.DeleteSnapshot(ctx, managedAppManagedResourceGroupID,
		resourceGroupName, snapshotID)
	if err != nil {
		return diag.Errorf("unable to delete snapshot (Managed Application %q) (Resource Group %q) (Correlation ID %q): %v",
			managedAppName,
			resourceGroupName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	err = crpClient.PollOperation(ctx, resp.Operation.ID, managedAppManagedResourceGroupID, managedAppName, 10)
	if err != nil {
		return diag.Errorf("unable to poll delete snapshot operation (Managed Application %q) (Resource Group %q) (Correlation ID %q): %v",
			managedAppName,
			resourceGroupName,
			meta.(*clients.Client).CorrelationRequestID,
			err,
		)
	}

	return nil
}

func populateSnapshotState(d *schema.ResourceData, snapshot *models.HashicorpCloudConsulamaAmaSnapshotProperties) diag.Diagnostics {
	if err := d.Set("state", snapshot.State); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("requested_at", snapshot.RequestedAt.String()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("finished_at", snapshot.FinishedAt.String()); err != nil {
		return diag.FromErr(err)
	}

	size, err := strconv.Atoi(snapshot.Size)
	if err != nil {
		return diag.Errorf("unable to convert string to int: %v", err)
	}
	if err := d.Set("size", size); err != nil {
		return diag.FromErr(err)
	}

	if snapshot.RestoredAt.String() != defaultRestoredAt {
		if err := d.Set("restored_at", snapshot.RestoredAt.String()); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
