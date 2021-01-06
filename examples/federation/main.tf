resource "azurerm_resource_group" "primary" {
  name     = "hcs-tf-federation-primary-rg"
  location = "westus2"
}

resource "hcs_cluster" "primary" {
  resource_group_name      = azurerm_resource_group.primary.name
  managed_application_name = "hcs-tf-federation-primary"
  email                    = "me@example.com"
  cluster_mode             = "production"
  min_consul_version       = "v1.9.0"
  vnet_cidr                = "172.25.16.0/24"
  consul_datacenter        = "hcs-tf-federation-example"
}

data "hcs_federation_token" "test" {
  resource_group_name      = hcs_cluster.primary.resource_group_name
  managed_application_name = hcs_cluster.primary.managed_application_name
}

resource "azurerm_resource_group" "secondary" {
  name     = "hcs-tf-federation-secondary-rg"
  location = "eastus"
}

resource "hcs_cluster" "secondary" {
  resource_group_name      = azurerm_resource_group.secondary.name
  managed_application_name = "hcs-tf-federation-secondary"
  email                    = "me@example.com"
  cluster_mode             = "production"
  min_consul_version       = "v1.9.0"
  vnet_cidr                = "172.25.17.0/24"
  consul_datacenter        = "hcs-tf-federation-secondary"
  consul_federation_token  = data.hcs_federation_token.test.token
}
