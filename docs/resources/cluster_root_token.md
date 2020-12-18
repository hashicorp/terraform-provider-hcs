---
page_title: "hcs_cluster_root_token Resource - terraform-provider-hcs"
subcategory: ""
description: |-
  The cluster root token resource is the token used to bootstrap the cluster's ACL system. Using this resource to create a new root token for an cluster resource will invalidate the consul root token accessor id and consul root token secret id properties of the cluster.
---

# Resource `hcs_cluster_root_token`

The cluster root token resource is the token used to bootstrap the cluster's ACL system. Using this resource to create a new root token for an cluster resource will invalidate the consul root token accessor id and consul root token secret id properties of the cluster.

## Example Usage

```terraform
// Note: creating a new root token for an hcs_cluster resource will invalidate the
// consul_root_token_accessor_id and consul_root_token_secret_id properties of the
// cluster.
resource "hcs_root_token" "new_token" {
  resource_group_name      = var.resource_group_name
  managed_application_name = var.managed_application_name
}
```

## Schema

### Required

- **managed_application_name** (String) The name of the HCS Azure Managed Application.
- **resource_group_name** (String) The name of the Resource Group in which the HCS Azure Managed Application belongs.

### Optional

- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **accessor_id** (String) The accessor ID of the root ACL token.
- **kubernetes_secret** (String, Sensitive) The root ACL token Base64 encoded in a Kubernetes secret.
- **secret_id** (String, Sensitive) The secret ID of the root ACL token.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


