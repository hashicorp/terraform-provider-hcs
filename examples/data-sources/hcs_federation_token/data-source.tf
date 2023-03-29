# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "hcs_federation_token" "default" {
  resource_group_name      = var.resource_group_name
  managed_application_name = var.managed_application_name
}