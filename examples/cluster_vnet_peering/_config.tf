terraform {
  required_providers {
    hcs = {
      // TODO: Update this to hashicorp/hcs when the provider is available on the registry
      source  = "localhost/providers/hcs"
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