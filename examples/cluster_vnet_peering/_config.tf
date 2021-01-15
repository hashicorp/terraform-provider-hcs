terraform {
  required_providers {
    hcs = {
      source  = "hashicorp/hcs"
      version = "~> 0.1.0"
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