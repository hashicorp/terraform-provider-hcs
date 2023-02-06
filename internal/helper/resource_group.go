// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package helper

import (
	"fmt"
	"strings"
)

// ParseResourceGroupNameFromID takes an Azure id string and parses the
// Azure Resource Group name. The Resource Group name can be the name of a Resource Group
// or a Managed Resource Group. Azure ids are of the form:
// /subscriptions/{guid}/resourceGroups/{resource-group-name}/{resource-provider-namespace}/{resource-type}/{resource-name}
func ParseResourceGroupNameFromID(id string) (string, error) {
	parts := strings.Split(strings.TrimPrefix(id, "/"), "/")
	if len(parts) < 4 || parts[2] != "resourceGroups" {
		return "", fmt.Errorf("unable to parse resource group name from id")
	}

	return parts[3], nil
}

// ParseNameFromID takes an Azure id string and parses the
// Resource name. The Resource name can be the name of a Resource Group
// or a Managed Resource Group. Azure ids are of the form:
// /subscriptions/{guid}/resourceGroups/{resource-group-name}/{resource-provider-namespace}/{resource-type}/{resource-name}
func ParseResourceNameFromID(id string) string {
	parts := strings.Split(id, "/")

	return parts[len(parts)-1]
}
