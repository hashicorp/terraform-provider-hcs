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