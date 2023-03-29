# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "hcs_agent_helm_config" "default" {
  resource_group_name      = var.resource_group_name
  managed_application_name = var.managed_application_name
  aks_cluster_name         = var.aks_cluster_name
  aks_resource_group       = var.aks_resource_group
  expose_gossip_ports      = var.expose_gossip_ports
}