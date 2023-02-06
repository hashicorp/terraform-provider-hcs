# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

// Note: Snapshots currently have a retention policy of 30 days. After that time, any Terraform
// state refresh will note that a new snapshot resource will be created.
resource "hcs_snapshot" "default" {
  resource_group_name      = var.resource_group_name
  managed_application_name = var.managed_application_name
  snapshot_name            = var.snapshot_name
}