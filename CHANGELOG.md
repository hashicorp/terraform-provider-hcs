## 0.5.0 (Unreleased)

IMPROVEMENTS:
* Add PlatformType to ListVersions call for `hcs_public_versions` datasource.

## 0.4.0 (September 08, 2021)

IMPROVEMENTS:
* `hcs_cluster` resource: Modified diff suppression of the `consul_version` field to ignore the field when the specified version is less than or equal to the clusters actual version.

## 0.3.0 (June 02, 2021)

IMPROVEMENTS:
* `hcs_cluster` resource: Added `audit_logging_enabled` and `audit_log_storage_container_url` to configure Consul audit logging to an Azure storage container. 
* `hcs_cluster` resource: Added `managed_identity_name` to schema for easy setup of the role assignment on the storage container for writing Consul audit logs.

BUG FIXES:
* Pass `cluster_mode` as a property when creating the `hcs_cluster` resource.

## 0.2.0 (March 01, 2021)

IMPROVEMENTS:
* Binary releases of this provider now include the darwin-arm64 platform.

## 0.1.1 (February 01, 2021)

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
