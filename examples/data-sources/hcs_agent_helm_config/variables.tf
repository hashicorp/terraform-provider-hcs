variable "resource_group_name" {
  type = string
}

variable "managed_application_name" {
  type = string
}

variable "aks_cluster_name" {
  type = string
}

variable "aks_resource_group" {
  type = string
}

variable "expose_gossip_ports" {
  type    = bool
  default = false
}
