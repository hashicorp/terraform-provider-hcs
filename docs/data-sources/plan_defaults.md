---
page_title: "hcs_plan_defaults Data Source - terraform-provider-hcs"
subcategory: ""
description: |-
  The plan defaults data source provides details about the current Azure Marketplace Plan defaults for the HCS offering. The plan defaults are useful when accepting the Azure Marketplace Agreement for the HCS Azure Managed Application.
---

# Data Source `hcs_plan_defaults`

The plan defaults data source provides details about the current Azure Marketplace Plan defaults for the HCS offering. The plan defaults are useful when accepting the Azure Marketplace Agreement for the HCS Azure Managed Application.

## Example Usage

```terraform
data "hcs_plan_defaults" "default" {}
```

## Schema

### Optional

- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **offer** (String) The name of the offer for the HCS Azure Managed Application.
- **plan_name** (String) The plan name for the HCS Azure Managed Application offer.
- **plan_version** (String) The plan version for the HCS Azure Managed Application offer.
- **publisher** (String) The publisher for the HCS Azure Managed Application offer.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


