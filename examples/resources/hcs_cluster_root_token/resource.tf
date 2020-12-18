// Note: creating a new root token for an hcs_cluster resource will invalidate the
// consul_root_token_accessor_id and consul_root_token_secret_id properties of the
// cluster.
resource "hcs_root_token" "new_token" {
  resource_group_name      = var.resource_group_name
  managed_application_name = var.managed_application_name
}