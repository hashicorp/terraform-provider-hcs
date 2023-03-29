# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "azurerm_resource_group" "example" {
  name     = "hcs-tf-root-token-example-rg"
  location = "westus2"
}

// The consul_root_token_accessor_id and consul_root_token_secret_id properties will
// no longer be valid after the new root token is created below
resource "hcs_cluster" "example" {
  resource_group_name      = azurerm_resource_group.example.name
  managed_application_name = "hcs-tf-root-token-example"
  email                    = "me@example.com"
  cluster_mode             = "production"
}

// Create a new ACL Root token
resource "hcs_cluster_root_token" "example" {
  resource_group_name      = azurerm_resource_group.example.name
  managed_application_name = hcs_cluster.example.managed_application_name
}
