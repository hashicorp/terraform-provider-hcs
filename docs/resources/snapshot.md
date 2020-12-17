---
page_title: "hcs_snapshot Resource - terraform-provider-hcs"
subcategory: ""
description: |-
  The snapshot resource allows users to manage Consul snapshots of an HCS cluster. Snapshots currently have a retention policy of 30 days.
---

# Resource `hcs_snapshot`

The snapshot resource allows users to manage Consul snapshots of an HCS cluster. Snapshots currently have a retention policy of 30 days.

## Example Usage

```terraform
// Note: Snapshots currently have a retention policy of 30 days. After that time, any Terraform
// state refresh will note that a new snapshot resource will be created.
resource "hcs_snapshot" "default" {
  resource_group_name      = var.resource_group_name
  managed_application_name = var.managed_application_name
  snapshot_name            = var.snapshot_name
}
```

## Schema

### Required

- **managed_application_name** (String, Required) The name of the HCS Azure Managed Application.
- **resource_group_name** (String, Required) The name of the Resource Group in which the HCS Azure Managed Application belongs.
- **snapshot_name** (String, Required) The name of the snapshot.

### Optional

- **id** (String, Optional) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **finished_at** (String, Read-only) Timestamp of when the snapshot was finished.
- **requested_at** (String, Read-only) Timestamp of when the snapshot was requested.
- **restored_at** (String, Read-only) Timestamp of when the snapshot was restored. If the snapshot has not been restored, this field will be blank.
- **size** (Number, Read-only) The size of the snapshot in bytes.
- **state** (String, Read-only) The state of the snapshot.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String, Optional)
- **default** (String, Optional)
- **delete** (String, Optional)
- **update** (String, Optional)


