package provider

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcs/internal/clients"
	"github.com/hashicorp/terraform-provider-hcs/internal/clients/hcs-ama-api-spec/models"
)

const (
	// defaultRestoredAt is the default string returned when a snapshot has not been restored
	defaultRestoredAt = "0001-01-01T00:00:00.000Z"
)

// snapshotCreateUpdateDeleteTimeoutDuration is the amount of time that can elapse
// before a snapshot operation should timeout.
var snapshotCreateUpdateDeleteTimeoutDuration = time.Minute * 15

// resourceSnapshot defines the snapshot resource schema and CRUD contexts.
func resourceSnapshot() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSnapshotCreate,
		ReadContext:   resourceSnapshotRead,
		UpdateContext: resourceSnapshotUpdate,
		DeleteContext: resourceSnapshotDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: &snapshotCreateUpdateDeleteTimeoutDuration,
			Update: &snapshotCreateUpdateDeleteTimeoutDuration,
			Delete: &snapshotCreateUpdateDeleteTimeoutDuration,
		},
		Schema: map[string]*schema.Schema{
			// Required inputs
			"resource_group_name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateResourceGroupName,
				ForceNew:         true,
			},
			"managed_application_name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateManagedAppName,
				ForceNew:         true,
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
	resourceGroupName := d.Get("resource_group_name").(string)
	managedAppName := d.Get("managed_application_name").(string)

	managedAppClient := meta.(*clients.Client).ManagedApplication
	app, err := managedAppClient.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		return diag.Errorf("failed to check for presence of existing HCS Cluster (Managed Application %q) (Resource Group %q): %+v", managedAppName, resourceGroupName, err)
	}
	if app.Response.StatusCode == 404 {
		// No managed application exists, so this snapshot should be removed from state
		log.Printf("[ERROR] no HCS Cluster found for (Managed Application %q) (Resource Group %q)", managedAppName, resourceGroupName)
		d.SetId("")
		return nil
	}

	managedAppManagedResourceGroupID := *app.ManagedResourceGroupID
	snapshotName := d.Get("snapshot_name").(string)

	crpClient := meta.(*clients.Client).CustomResourceProvider
	resp, err := crpClient.CreateSnapshot(ctx, managedAppManagedResourceGroupID,
		resourceGroupName, snapshotName)
	if err != nil {
		return diag.Errorf("failed to create snapshot for managed app: %+v", err)
	}

	d.SetId(resp.SnapshotID)

	err = crpClient.PollOperation(ctx, resp.Operation.ID, managedAppManagedResourceGroupID, managedAppName, 10)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error polling operation: %+v", err))
	}

	return resourceSnapshotRead(ctx, d, meta)
}

func resourceSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resourceGroupName := d.Get("resource_group_name").(string)
	managedAppName := d.Get("managed_application_name").(string)

	managedAppClient := meta.(*clients.Client).ManagedApplication
	app, err := managedAppClient.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		return diag.Errorf("failed to check for presence of existing HCS Cluster (Managed Application %q) (Resource Group %q): %+v", managedAppName, resourceGroupName, err)
	}
	if app.Response.StatusCode == 404 {
		// No managed application exists, so this snapshot should be removed from state
		log.Printf("[ERROR] no HCS Cluster found for (Managed Application %q) (Resource Group %q)", managedAppName, resourceGroupName)
		d.SetId("")
		return nil
	}

	managedAppManagedResourceGroupID := *app.ManagedResourceGroupID
	snapshotID := d.Id()

	crpClient := meta.(*clients.Client).CustomResourceProvider
	resp, err := crpClient.GetSnapshot(ctx, managedAppManagedResourceGroupID,
		resourceGroupName, snapshotID)

	if err != nil {
		azErr, ok := err.(*azure.RequestError)
		if !ok {
			return diag.Errorf("failed to get snapshot for managed app: %+v", err)
		}

		if azErr.StatusCode == 404 {
			log.Printf("[ERROR] snapshot not found. the retention policy for snapshots is 30 days and " +
				"this snapshot may have been deleted, if you leave the snapshot resource " +
				"in your plan, a new snapshot will be created")
			d.SetId("")
			return nil
		}
	}

	if diag := populateSnapshotState(d, resp.Snapshot); diag != nil {
		return diag
	}

	return nil
}

func resourceSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resourceGroupName := d.Get("resource_group_name").(string)
	managedAppName := d.Get("managed_application_name").(string)

	managedAppClient := meta.(*clients.Client).ManagedApplication
	app, err := managedAppClient.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		return diag.Errorf("failed to check for presence of existing HCS Cluster (Managed Application %q) (Resource Group %q): %+v", managedAppName, resourceGroupName, err)
	}
	if app.Response.StatusCode == 404 {
		// No managed application exists, so this snapshot should be removed from state
		log.Printf("[ERROR] no HCS Cluster found for (Managed Application %q) (Resource Group %q)", managedAppName, resourceGroupName)
		d.SetId("")
		return nil
	}

	managedResourceGroupID := *app.ManagedResourceGroupID
	snapshotName := d.Get("snapshot_name").(string)
	snapshotID := d.Id()

	crpClient := meta.(*clients.Client).CustomResourceProvider
	resp, err := crpClient.RenameSnapshot(ctx, managedResourceGroupID, resourceGroupName, snapshotID, snapshotName)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to rename snapshot for managed app: %s", err))
	}

	if diag := populateSnapshotState(d, resp.Snapshot); diag != nil {
		return diag
	}

	return nil
}

func resourceSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resourceGroupName := d.Get("resource_group_name").(string)
	managedAppName := d.Get("managed_application_name").(string)

	managedAppClient := meta.(*clients.Client).ManagedApplication
	app, err := managedAppClient.Get(ctx, resourceGroupName, managedAppName)
	if err != nil {
		return diag.Errorf("failed to check for presence of existing HCS Cluster (Managed Application %q) (Resource Group %q): %+v", managedAppName, resourceGroupName, err)
	}
	if app.Response.StatusCode == 404 {
		// No managed application exists, so this snapshot should be removed from state
		log.Printf("[ERROR] no HCS Cluster found for (Managed Application %q) (Resource Group %q)", managedAppName, resourceGroupName)
		d.SetId("")
		return nil
	}

	managedAppManagedResourceGroupID := *app.ManagedResourceGroupID
	snapshotID := d.Id()

	crpClient := meta.(*clients.Client).CustomResourceProvider
	resp, err := crpClient.DeleteSnapshot(ctx, managedAppManagedResourceGroupID,
		resourceGroupName, snapshotID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete snapshot for managed app: %s", err))
	}

	err = crpClient.PollOperation(ctx, resp.Operation.ID, managedAppManagedResourceGroupID, managedAppName, 10)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error polling operation: %+v", err))
	}

	d.SetId("")
	return nil
}

func populateSnapshotState(d *schema.ResourceData, snapshot *models.HashicorpCloudConsulamaAmaSnapshotProperties) diag.Diagnostics {
	d.Set("state", snapshot.State)
	d.Set("requested_at", snapshot.RequestedAt.String())
	d.Set("finished_at", snapshot.FinishedAt.String())

	size, err := strconv.Atoi(snapshot.Size)
	if err != nil {
		return diag.FromErr(fmt.Errorf("[ERROR] Error converting string to int: %+v", err))
	}
	d.Set("size", size)

	if snapshot.RestoredAt.String() != defaultRestoredAt {
		d.Set("restored_at", snapshot.RestoredAt.String())
	}

	return nil
}
