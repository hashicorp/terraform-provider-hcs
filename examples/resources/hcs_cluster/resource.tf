data "hcs_consul_versions" "default" {}

data "hcs_plan_defaults" "default" {}

resource "hcs_cluster" "example" {
  resource_group_name      = var.resource_group_name
  managed_application_name = var.managed_application_name
  email                    = var.email
  cluster_mode             = var.cluster_mode
  vnet_cidr                = var.vnet_cidr
  consul_version           = data.hcs_consul_versions.default.recommended
  location                 = var.location
  plan_name                = data.hcs_plan_defaults.default.plan_name
}