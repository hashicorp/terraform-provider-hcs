---
page_title: "hcs_cluster Data Source - terraform-provider-hcs"
subcategory: ""
description: |-
  The cluster data source provides information about an existing HCS cluster.
---

# Data Source `hcs_cluster`

The cluster data source provides information about an existing HCS cluster.

## Example Usage

```terraform
data "hcs_cluster" "default" {
  resource_group_name      = var.resource_group_name
  managed_application_name = var.managed_application_name
  // cluster_name can be omitted if the cluster's name matches its managed_application_name
  cluster_name = var.cluster_name
}
```

## Schema

### Required

- **managed_application_name** (String) The name of the HCS Azure Managed Application.
- **resource_group_name** (String) The name of the Resource Group in which the HCS Azure Managed Application belongs.

### Optional

- **cluster_name** (String) The name of the cluster Managed Resource. If not specified, it is defaulted to the value of `managed_application_name`.
- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **blob_container_name** (String) The name of the Blob Container in which cluster data is persisted.
- **cluster_mode** (String) The mode of the cluster ('Development' or 'Production'). Development clusters only have a single Consul server. Production clusters are fully supported, full featured, and deploy with a minimum of three hosts.
- **consul_automatic_upgrades** (Boolean) Denotes that automatic Consul upgrades are enabled.
- **consul_ca_file** (String) The cluster CA file encoded as a Base64 string.
- **consul_cluster_id** (String) The cluster ID.
- **consul_config_file** (String) The cluster config encoded as a Base64 string.
- **consul_connect** (Boolean) Denotes that Consul connect is enabled.
- **consul_datacenter** (String) The Consul data center name of the cluster.
- **consul_external_endpoint** (Boolean) Denotes that the cluster has an external endpoint for the Consul UI.
- **consul_external_endpoint_url** (String) The public URL for the Consul UI. This will be empty if `consul_external_endpoint` is `true`.
- **consul_federation_token** (String) The token used to join a federation of Consul clusters. If the cluster is not part of a federation, this field will be empty.
- **consul_private_endpoint_url** (String) The private URL for the Consul UI.
- **consul_snapshot_interval** (String) The Consul snapshot interval.
- **consul_snapshot_retention** (String) The retention policy for Consul snapshots.
- **consul_version** (String) The Consul version of the cluster.
- **email** (String) The contact email for the primary owner of the cluster.
- **location** (String) The Azure region that the cluster is deployed to.
- **managed_application_id** (String) The ID of the Managed Application.
- **managed_resource_group_name** (String) The name of the Managed Resource Group in which the cluster resources belong.
- **plan_name** (String) The name of the Azure Marketplace HCS plan for the cluster.
- **state** (String) The state of the cluster.
- **storage_account_name** (String) The name of the Storage Account in which cluster data is persisted.
- **storage_account_resource_group** (String) The name of the Storage Account's Resource Group.
- **vnet_cidr** (String) The VNET CIDR range of the Consul cluster.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


