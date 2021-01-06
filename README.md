HashiCorp Consul Service on Azure (HCS) Terraform Provider
==================

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.12.x
-	[Go](https://golang.org/doc/install) >= 1.14

Building The Provider
---------------------

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the `make dev` command

Adding Dependencies
---------------------

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate the latest models for the HCS Custom Resource Provider actions, run:
```
make generate-hcs-ama-api-spec-models
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```
 
 Generating Docs
 ----------------------
 
 From the root of the repo run:
 
 ```
 go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
 ```

 
Using the provider
----------------------

Please see the docs for details about a particular resource. 
Below is a complex example that leverages the Azure Terraform provider and creates a federation of two HCS clusters.
```hcl
// Configure the provider
provider "hcs" {}

provider "azurerm" {
  features {}
}

// If you have not already done so, accept the HCS Marketplace agreement
data "hcs_plan_defaults" "hcs_plan" {}

resource "azurerm_marketplace_agreement" "hcs_marketplace_agreement" {
  publisher = data.hcs_plan_defaults.hcs_plan.publisher
  offer     = data.hcs_plan_defaults.hcs_plan.offer
  plan      = data.hcs_plan_defaults.hcs_plan.plan_name
}

// Create the Resource Group for the primary cluster
resource "azurerm_resource_group" "primary" {
  name     = "hcs-tf-federation-primary-rg"
  location = "westus2"
}

// Create the primary cluster
resource "hcs_cluster" "primary" {
  resource_group_name      = azurerm_resource_group.primary.name
  managed_application_name = "hcs-tf-federation-primary"
  email                    = "me@example.com"
  cluster_mode             = "production"
  min_consul_version       = "v1.9.0"
  vnet_cidr                = "172.25.16.0/24"
  consul_datacenter        = "hcs-tf-federation-example"
}

// Create a federation token
data "hcs_federation_token" "fed" {
  resource_group_name      = hcs_cluster.primary.resource_group_name
  managed_application_name = hcs_cluster.primary.managed_application_name
}

// Create the Resource Group for the secondary cluster
resource "azurerm_resource_group" "secondary" {
  name     = "hcs-tf-federation-secondary-rg"
  location = "eastus"
}

// Create the secondary cluster using the federation token from above
resource "hcs_cluster" "secondary" {
  resource_group_name      = azurerm_resource_group.secondary.name
  managed_application_name = "hcs-tf-federation-secondary"
  email                    = "me@example.com"
  cluster_mode             = "production"
  min_consul_version       = "v1.9.0"
  vnet_cidr                = "172.25.17.0/24"
  consul_datacenter        = "hcs-tf-federation-secondary"
  consul_federation_token  = data.hcs_federation_token.fed.token
}
```