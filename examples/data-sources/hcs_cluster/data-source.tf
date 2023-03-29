# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "hcs_cluster" "default" {
  resource_group_name      = var.resource_group_name
  managed_application_name = var.managed_application_name
  // cluster_name can be omitted if the cluster's name matches its managed_application_name
  cluster_name = var.cluster_name
}