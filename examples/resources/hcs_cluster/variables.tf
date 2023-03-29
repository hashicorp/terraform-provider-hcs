# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

variable "resource_group_name" {
  type = string
}

variable "managed_application_name" {
  type = string
}

variable "email" {
  type = string
}

variable "cluster_mode" {
  type    = string
  default = "Production"
}

variable "vnet_cidr" {
  type = string
}

variable "location" {
  type = string
}