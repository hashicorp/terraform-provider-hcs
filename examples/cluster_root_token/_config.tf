# Copyright (c) HashiCorp, Inc.
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