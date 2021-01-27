## 0.1.1 (Unreleased)

IMPROVEMENTS:
* `hcs_cluster` resource: `min_consul_version` now can handle semver versions with metadata or a prerelease.
* `hcs_cluster` resource: Create timeout increased to 60 minutes.
* `hcs_cluster` data source: Added `vnet_id`, `vnet_name`, and `vnet_resource_group_name` to schema. 

## 0.1.0 (January 15, 2021)

FEATURES:
* **New resource** `hcs_cluster`.
* **New data source** `hcs_cluster`.
* **New resource** `hcs_cluster_root_token`.
* **New resource** `hcs_snapshot`.
* **New data source** `hcs_agent_helm_config`.
* **New data source** `hcs_agent_kubernetes_secret`.
* **New data source** `hcs_consul_versions`.
* **New data source** `hcs_federation_token`.
* **New data source** `hcs_plan_defaults`.
