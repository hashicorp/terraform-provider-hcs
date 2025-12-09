# Copyright IBM Corp. 2020, 2025
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_providers {
    hcs = {
      source  = "hashicorp/hcs"
      version = "~> 0.1"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 2.39"
    }
  }
}

provider "hcs" {}

provider "azurerm" {
  features {}
}