---
page_title: "hcs_cluster Resource - terraform-provider-hcs"
subcategory: ""
description: |-
  The cluster resource allows you to manage an HCS Azure Managed Application.
---

# Resource `hcs_cluster`

The cluster resource allows you to manage an HCS Azure Managed Application.

## Example Usage

```terraform
data "hcs_consul_versions" "default" {}

data "hcs_plan_defaults" "default" {}

resource "hcs_cluster" "example" {
  resource_group_name      = var.resource_group_name
  managed_application_name = var.managed_application_name
  email                    = var.email
  cluster_mode             = var.cluster_mode
  vnet_cidr                = var.vnet_cidr
  consul_version           = data.hcs_consul_versions.default.recommended
  location                 = var.location
  plan_name                = data.hcs_plan_defaults.default.plan_name
}
```

## Schema

### Required

- **cluster_mode** (String, Required) The mode of the cluster ('Development' or 'Production'). Development clusters only have a single Consul server node. Production clusters deploy with a minimum of three nodes.
- **email** (String, Required) The contact email for the primary owner of the cluster.
- **managed_application_name** (String, Required) The name of the HCS Azure Managed Application.
- **resource_group_name** (String, Required) The name of the Resource Group in which the HCS Azure Managed Application belongs.

### Optional

- **cluster_name** (String, Optional) The name of the cluster Managed Resource. If not specified, it is defaulted to the value of `managed_application_name`.
- **consul_datacenter** (String, Optional) The Consul data center name of the cluster. If not specified, it is defaulted to the value of `managed_application_name`.
- **consul_external_endpoint** (Boolean, Optional) Denotes that the cluster has an external endpoint for the Consul UI.
- **consul_federation_token** (String, Optional) The token used to join a federation of Consul clusters. If the cluster is not part of a federation, this field will be empty.
- **consul_version** (String, Optional) The Consul version of the cluster. If not specified, it is defaulted to the version that is currently recommended by HCS.
- **id** (String, Optional) The ID of this resource.
- **location** (String, Optional) The Azure region that the cluster is deployed to. If not specified, it is defaulted to the region of the Resource Group the Managed Application belongs to.
- **managed_resource_group_name** (String, Optional) The name of the Managed Resource Group in which the cluster resources belong. If not specified, it is defaulted to the value of `managed_application_name` with 'mrg-' prepended.
- **plan_name** (String, Optional) The name of the Azure Marketplace HCS plan for the cluster. If not specified, it will default to the current HCS default plan (see the `hcs_plan_defaults` data source).
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- **vnet_cidr** (String, Optional) The VNET CIDR range of the Consul cluster.

### Read-only

- **blob_container_name** (String, Read-only) The name of the Blob Container in which cluster data is persisted.
- **consul_automatic_upgrades** (Boolean, Read-only) Denotes that automatic Consul upgrades are enabled.
- **consul_ca_file** (String, Read-only) The cluster CA file encoded as a Base64 string.
- **consul_cluster_id** (String, Read-only) The cluster ID.
- **consul_config_file** (String, Read-only) The cluster config encoded as a Base64 string.
- **consul_connect** (Boolean, Read-only) Denotes that Consul connect is enabled.
- **consul_external_endpoint_url** (String, Read-only) The public URL for the Consul UI. This will be empty if `consul_external_endpoint` is `true`.
- **consul_private_endpoint_url** (String, Read-only) The private URL for the Consul UI.
- **consul_root_token_accessor_id** (String, Read-only) The accessor ID of the root ACL token that is generated upon cluster creation. If a new root token is generated using the `hcs_cluster_root_token` resource, this field is no longer valid.
- **consul_root_token_secret_id** (String, Read-only) The secret ID of the root ACL token that is generated upon cluster creation. If a new root token is generated using the `hcs_cluster_root_token` resource, this field is no longer valid.
- **consul_snapshot_interval** (String, Read-only) The Consul snapshot interval.
- **consul_snapshot_retention** (String, Read-only) The retention policy for Consul snapshots.
- **managed_application_id** (String, Read-only) The ID of the Managed Application.
- **state** (String, Read-only) The state of the cluster.
- **storage_account_name** (String, Read-only) The name of the Storage Account in which cluster data is persisted.
- **storage_account_resource_group** (String, Read-only) The name of the Storage Account's Resource Group.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **create** (String, Optional)
- **default** (String, Optional)
- **delete** (String, Optional)
- **update** (String, Optional)

## Import

Import is supported using the following syntax:

```shell
# The import ID is {Managed Application ID}:{Cluster Name}
terraform import hcs_cluster.example /subscriptions/1234-5678-91011-1213-141516/resourceGroups/hcs-tf-example/providers/Microsoft.Solutions/applications/hcs-tf-example:hcs-tf-example
```
