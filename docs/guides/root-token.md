---
subcategory: ""
page_title: "Create a new ACL root token - HCS Provider"
description: |-
    An example of creating a new ACL root token.
---

# Create a new Consul ACL root token

Creating a new root token is useful if your HCS cluster has been imported into Terraform
or is managed outside of Terraform. It is important to note that when creating a new root token,
the existing root token will be invalidated.

```terraform
resource "azurerm_resource_group" "example" {
  name     = "hcs-tf-root-token-example-rg"
  location = "westus2"
}

// The consul_root_token_accessor_id and consul_root_token_secret_id properties will
// no longer be valid after the new root token is created below
resource "hcs_cluster" "example" {
  resource_group_name      = azurerm_resource_group.example.name
  managed_application_name = "hcs-tf-root-token-example"
  email                    = "me@example.com"
  cluster_mode             = "production"
}

// Create a new ACL Root token
resource "hcs_cluster_root_token" "example" {
  resource_group_name      = azurerm_resource_group.example.name
  managed_application_name = hcs_cluster.example.managed_application_name
}
```
