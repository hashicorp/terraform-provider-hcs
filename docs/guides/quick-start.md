---
subcategory: ""
page_title: "Create an HCS cluster - HCS Provider"
description: |-
    An example of creating an HCS cluster with optional fields defaulted.
---

# Create an HCS cluster using the `hcs_cluster` resource

```terraform
resource "azurerm_resource_group" "example" {
  name     = "hcs-tf-example-tags"
  location = "westus2"
}

resource "hcs_cluster" "example" {
  resource_group_name      = azurerm_resource_group.example.name
  managed_application_name = "hcs-tf-example-tags"
  email                    = "me@example.com"
  cluster_mode             = "production"
  tags = {
    foo = "bar"
  }
}
```
