---
page_title: "hcs_consul_versions Data Source - terraform-provider-hcs"
subcategory: ""
description: |-
  The Consul versions data source provides the Consul versions supported by HCS.
---

# Data Source `hcs_consul_versions`

The Consul versions data source provides the Consul versions supported by HCS.

## Example Usage

```terraform
data "hcs_consul_versions" "default" {}
```

## Schema

### Optional

- **id** (String) The ID of this resource.
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-only

- **available** (List of String) The Consul versions available on HCS.
- **preview** (List of String) The preview versions of Consul available on HCS.
- **recommended** (String) The recommended Consul version for HCS clusters.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)

