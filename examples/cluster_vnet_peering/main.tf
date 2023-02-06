# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "azurerm_resource_group" "example" {
  name     = "hcs-tf-example"
  location = "westus2"
}

resource "hcs_cluster" "example" {
  resource_group_name      = azurerm_resource_group.example.name
  managed_application_name = "hcs-tf-example"
  email                    = "me@example.com"
  cluster_mode             = "production"
}

resource "azurerm_virtual_network" "example" {
  name                = "peer-network"
  resource_group_name = azurerm_resource_group.example.name
  address_space       = ["10.0.2.0/24"]
  location            = "westus2"
}

resource "azurerm_virtual_network_peering" "cluster-to-network" {
  name                      = "cluster-to-network"
  resource_group_name       = hcs_cluster.example.vnet_resource_group_name
  virtual_network_name      = hcs_cluster.example.vnet_name
  remote_virtual_network_id = azurerm_virtual_network.example.id
}

resource "azurerm_virtual_network_peering" "network-to-cluster" {
  name                      = "network-to-cluster"
  resource_group_name       = azurerm_resource_group.example.name
  virtual_network_name      = azurerm_virtual_network.example.name
  remote_virtual_network_id = hcs_cluster.example.vnet_id
}