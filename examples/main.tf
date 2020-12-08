terraform {
  required_providers {
    hcs = {
      source  = "unreleased.hashicorp.com/hashicorp/hcs"
      version = "0.0.1"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 2.39.0"
    }
  }
}

provider "hcs" {}

provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "example" {
  name     = "hcs-tf-example"
  location = "westus2"
}

resource "hcs_cluster" "test" {
  resource_group_name      = azurerm_resource_group.example.name
  managed_application_name = "hcs-tf-example"
  email                    = "me@example.com"
  cluster_mode             = "production"
}
