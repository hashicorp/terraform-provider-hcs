---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "hcs_agent_helm_config Data Source - terraform-provider-hcs"
subcategory: ""
description: |-
  The agent Helm config data source provides Helm values for a Consul agent running in Kubernetes.
---

# hcs_agent_helm_config (Data Source)

The agent Helm config data source provides Helm values for a Consul agent running in Kubernetes.

## Example Usage

```terraform
data "hcs_agent_helm_config" "default" {
  resource_group_name      = var.resource_group_name
  managed_application_name = var.managed_application_name
  aks_cluster_name         = var.aks_cluster_name
  aks_resource_group       = var.aks_resource_group
  expose_gossip_ports      = var.expose_gossip_ports
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **aks_cluster_name** (String) The name of the AKS cluster that will consume the Helm config.
- **managed_application_name** (String) The name of the HCS Azure Managed Application.
- **resource_group_name** (String) The name of the Resource Group in which the HCS Azure Managed Application belongs.

### Optional

- **aks_resource_group** (String) The resource group name of the AKS cluster that will consume the Helm config. If not specified, it is defaulted to the value of `resource_group_name`.
- **expose_gossip_ports** (Boolean) Denotes that the gossip ports should be exposed. Defaults to `false`.
- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- **config** (String) The agent Helm config.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


