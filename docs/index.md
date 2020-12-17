---
layout: ""
page_title: "Provider: HCS (HashiCorp Consul Service on Azure)"
description: |-
  The HCS provider provides resources to manage [HashiCorp Consul Service on Azure](https://www.hashicorp.com/products/consul/service-on-azure) (HCS) clusters.
---

# HashiCorp Consul Service on Azure (HCS) Provider

The HCS provider provides resources to manage [HashiCorp Consul Service on Azure](https://www.hashicorp.com/products/consul/service-on-azure) (HCS) clusters.

## Authenticating to Azure
The HCS provider supports the same authentication methods as the [Azure provider](https://registry.terraform.io/providers/hashicorp/azurerm/2.40.0/docs#authenticating-to-azure).

## Example Usage

```terraform
// Configure the provider
provider "hcs" {}

// The HCS provider works great with resources provisioned via the Azure provider
provider "azurerm" {
  features {}
}

// Create an Azure Resource Group
resource "azurerm_resource_group" "example" {
  name     = "hcs-tf-example-rg"
  location = "westus2"
}

// Create an HCS cluster Azure Managed Application.
resource "hcs_cluster" "example" {
  resource_group_name      = azurerm_resource_group.example.name
  managed_application_name = "hcs-tf-example"
  email                    = "me@example.com"
  cluster_mode             = "production"
}
```

## Schema

### Optional

- **azure_client_certificate_password** (String, Optional) The password associated with the Azure Client Certificate. For use when authenticating as a Service Principal using a Client Certificate
- **azure_client_certificate_path** (String, Optional) The path to the Azure Client Certificate associated with the Service Principal for use when authenticating as a Service Principal using a Client Certificate.
- **azure_client_id** (String, Optional) The Azure Client ID which should be used.
- **azure_client_secret** (String, Optional) The Azure Client Secret which should be used. For use when authenticating as a Service Principal using a Client Secret.
- **azure_environment** (String, Optional) The Azure Cloud Environment which should be used. Possible values are public, usgovernment, german, and china. Defaults to public.
- **azure_metadata_host** (String, Optional) The hostname which should be used for the Azure Metadata Service.
- **azure_msi_endpoint** (String, Optional) The path to a custom endpoint for Azure Managed Service Identity - in most circumstances this should be detected automatically.
- **azure_subscription_id** (String, Optional) The Azure Subscription ID which should be used.
- **azure_tenant_id** (String, Optional) The Azure Tenant ID which should be used.
- **azure_use_msi** (Boolean, Optional) Allowed Azure Managed Service Identity be used for Authentication.
- **hcp_api_domain** (String, Optional) The HashiCorp Cloud Platform API domain.
- **hcs_marketplace_product_name** (String, Optional) The HashiCorp Consul Service product name on the Azure marketplace.