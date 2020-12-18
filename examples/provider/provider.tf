// Configure the provider
provider "hcs" {}

// The HCS provider works great with resources provisioned via the Azure provider
provider "azurerm" {
  features {}
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