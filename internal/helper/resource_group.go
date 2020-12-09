package helper

import (
	"fmt"
	"strings"
)

// ParseResourceGroupNameFromID takes an Azure id string and parses the
// Azure Resource Group name. The Resource Group name can be the name of a Resource Group
// or a Managed Resource Group. Azure ids are of the form:
// /subscriptions/{GUID}/resourceGroups/{RESOURCE_GROUP_NAME}/...
func ParseResourceGroupNameFromID(id string) (string, error) {
	parts := strings.Split(strings.TrimPrefix(id, "/"), "/")
	if len(parts) < 4 || parts[2] != "resourceGroups" {
		return "", fmt.Errorf("unable to parse resource group name from id")
	}

	return parts[3], nil
}
