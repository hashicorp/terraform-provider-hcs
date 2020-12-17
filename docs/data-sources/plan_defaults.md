---
page_title: "hcs_plan_defaults Data Source - terraform-provider-hcs"
subcategory: ""
description: |-
  The plan defaults data source is useful for accepting the Azure Marketplace Agreement for the HCS Managed Application.
---

# Data Source `hcs_plan_defaults`

The plan defaults data source is useful for accepting the Azure Marketplace Agreement for the HCS Managed Application.

## Example Usage

```terraform
data "hcs_plan_defaults" "default" {}
```

## Schema

### Optional

- **id** (String, Optional) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **offer** (String, Read-only) The name of the offer for the HCS Managed Application.
- **plan_name** (String, Read-only) The plan name for the HCS Managed Application offer.
- **plan_version** (String, Read-only) The plan version for the HCS Managed Application offer.
- **publisher** (String, Read-only) The publisher for the HCS Managed Application offer.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String, Optional)


