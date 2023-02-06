// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccResourceScaffolding(t *testing.T) {
	t.Skip("resource not yet implemented, remove this once you add your own code")

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceScaffolding,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"scaffolding_resource.foo", "sample_attribute", regexp.MustCompile("^ba")),
				),
			},
		},
	})
}

const testAccResourceScaffolding = `
resource "scaffolding_resource" "foo" {
  sample_attribute = "bar"
}
`

func Test_validateClusterImportString(t *testing.T) {
	tcs := []struct {
		name         string
		importStr    string
		managedAppId string
		clusterName  string
		err          error
	}{
		{
			name:         "valid import string",
			importStr:    "/subscriptions/dadbabad-d00d-dada-baad-d00daaaaaaaa/resourceGroups/resource-group/providers/Microsoft.Solutions/applications/app1000:clusterName",
			managedAppId: "/subscriptions/dadbabad-d00d-dada-baad-d00daaaaaaaa/resourceGroups/resource-group/providers/Microsoft.Solutions/applications/app1000",
			clusterName:  "clusterName",
		},
		{
			name:      "invalid no colon",
			importStr: "invalid-no-colon",
			err:       fmt.Errorf("import id string must be of format `managed_application_id:cluster_name`; id string: %s does not contain `:`", "invalid-no-colon"),
		},
		{
			name:      "invalid more than one colon",
			importStr: "invalid:multiple:colon",
			err:       fmt.Errorf("import id string must be of format `managed_application_id:cluster_name`; id string: %s contains more than one `:`", "invalid:multiple:colon"),
		},
		{
			name:      "invalid empty id",
			importStr: ":cluster_name",
			err:       fmt.Errorf("import id string must be of format `managed_application_id:cluster_name`; id string: %s has empty string to left of `:`", ":cluster_name"),
		},
		{
			name:      "invalid empty cluster name",
			importStr: "id:",
			err:       fmt.Errorf("import id string must be of format `managed_application_id:cluster_name`; id string: %s has empty string to right of `:`", "id:"),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			id, name, err := validateClusterImportString(tc.importStr)
			if tc.err != nil {
				r.Equal(tc.err, err)
			} else {
				r.Equal(tc.managedAppId, id)
				r.Equal(tc.clusterName, name)
			}
		})
	}
}
