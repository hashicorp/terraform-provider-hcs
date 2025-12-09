# Copyright IBM Corp. 2020, 2025
# SPDX-License-Identifier: MPL-2.0

variable "resource_group_name" {
  type = string
}

variable "managed_application_name" {
  type = string
}

variable "aks_cluster_name" {
  type = string
}

variable "aks_resource_group" {
  type = string
}

variable "expose_gossip_ports" {
  type    = bool
  default = false
}
