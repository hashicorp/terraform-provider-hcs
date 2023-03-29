# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

// Configure the provider
provider "hcs" {}

// The HCS provider works great with resources provisioned via the Azure provider
provider "azurerm" {
  features {}
}

// If you have not already done so, accept the HCS Marketplace agreement
data "hcs_plan_defaults" "hcs_plan" {}

resource "azurerm_marketplace_agreement" "hcs_marketplace_agreement" {
  publisher = data.hcs_plan_defaults.hcs_plan.publisher
  offer     = data.hcs_plan_defaults.hcs_plan.offer
  plan      = data.hcs_plan_defaults.hcs_plan.plan_name
}

// Create an Azure Resource Group
resource "azurerm_resource_group" "example" {
  name     = "hcs-tf-example-rg"
  location = "westus2"
}

// Create an HCS cluster Azure Managed Application.
resource "hcs_cluster" "example" {
  resource_group_name      = azurerm_resource_group.example.name
  managed_application_name = "hcs-tf-example"
  email                    = "me@example.com"
  cluster_mode             = "production"
}