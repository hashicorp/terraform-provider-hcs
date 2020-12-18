---
subcategory: ""
page_title: "Accept the Azure Marketplace Agreement for HCS - HCS Provider"
description: |-
    An example of accepting the Azure Marketplace Agreement for the managed HCS offering.
---

# Accept the Azure Marketplace Agreement for HCS

```terraform
data "hcs_plan_defaults" "example" {}

resource "azurerm_marketplace_agreement" "example" {
  publisher = data.hcs_plan_defaults.example.publisher
  offer     = data.hcs_plan_defaults.example.offer
  plan      = data.hcs_plan_defaults.example.plan_name
}

resource "azurerm_resource_group" "example" {
  name     = "hcs-tf-plan-example"
  location = "westus2"
}

resource "hcs_cluster" "example" {
  resource_group_name      = azurerm_resource_group.example.name
  managed_application_name = "hcs-tf-plan-example"
  email                    = "me@example.com"
  cluster_mode             = "production"
  plan_name                = azurerm_marketplace_agreement.example.plan
}
```
