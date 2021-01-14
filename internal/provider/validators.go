package provider

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-hcs/internal/helper"
)

// validateStringNotEmpty ensures a given string is non-empty.
func validateStringNotEmpty(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	if v.(string) == "" {
		msg := "cannot be empty"
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg,
			AttributePath: path,
		})
	}

	return diagnostics
}

// validateResourceGroupName validates a resource group name string.
// Adapted from the azurerm provider
// https://github.com/terraform-providers/terraform-provider-azurerm/blob/8f32ad645888ee00a24ad7c739a8703222e13913/azurerm/helpers/azure/resource_group.go#L77
func validateResourceGroupName(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	value := v.(string)
	if len(value) > 90 {
		msg := "may not exceed 90 characters in length"
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg,
			AttributePath: path,
		})
	}

	if strings.HasSuffix(value, ".") {
		msg := "may not end with a period"
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg,
			AttributePath: path,
		})
	}

	// regex pulled from https://docs.microsoft.com/en-us/rest/api/resources/resourcegroups/createorupdate
	if matched := regexp.MustCompile(`^[-\w._()]+$`).Match([]byte(value)); !matched {
		msg := "may only contain alphanumeric characters, dash, underscores, parentheses and periods"
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg,
			AttributePath: path,
		})
	}

	return diagnostics
}

// validateSlugID validates that the string value matches the HCS requirements for
// a user-settable slug, as well as the Azure requirements for a Managed Application name.
func validateSlugID(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	// HCS supports a max of 36 chars for the cluster name which is defaulted to
	// the value of of the Managed App name so we must enforce a max of 36 even though
	// Azure supports a max of 64 chars for the Managed App name
	if !regexp.MustCompile(`^[-\da-zA-Z]{3,36}$`).MatchString(v.(string)) {
		msg := "must be between 3 and 36 characters in length and contains only letters, numbers or hyphens"
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg,
			AttributePath: path,
		})
	}

	return diagnostics
}

// validateStringInSlice returns a func which ensures the string value is a contained in the given slice.
// If ignoreCase is set the strings will be compared as lowercase.
// Adapted from terraform-plugin-sdk validate.StringInSlice
// https://github.com/hashicorp/terraform-plugin-sdk/blob/98ba036fe5895876219331532140d3d8cf239594/helper/validation/strings.go#L132
func validateStringInSlice(valid []string, ignoreCase bool) schema.SchemaValidateDiagFunc {
	return func(v interface{}, path cty.Path) diag.Diagnostics {
		var diagnostics diag.Diagnostics

		value := v.(string)

		for _, validString := range valid {
			if v == validString || (ignoreCase && strings.ToLower(value) == strings.ToLower(validString)) {
				return diagnostics
			}
		}

		msg := fmt.Sprintf("expected %s to be one of %v", value, valid)
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg,
			AttributePath: path,
		})
		return diagnostics
	}
}

// validateCIDR ensures that the provided string is a valid CIDR.
func validateCIDR(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	if _, _, err := net.ParseCIDR(v.(string)); err != nil {
		msg := "expected a valid CIDR"
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg,
			AttributePath: path,
		})
	}

	return diagnostics
}

// validateSemVer ensures a specified string is a SemVer.
func validateSemVer(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	if !regexp.MustCompile(`^v?\d+.\d+.\d+$`).MatchString(v.(string)) {
		msg := "must be a valid semver"
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg,
			AttributePath: path,
		})
	}

	return diagnostics
}

// validateManagedAppName ensures that the provided value is a string and follows
// the Azure convention for a managed app name.
func validateManagedAppName(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	value, ok := v.(string)
	if !ok {
		msg := "must be of type: string"
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg,
			AttributePath: path,
		})
		return diagnostics
	}

	// HCS supports a max of 36 chars for the cluster name which is defaulted to
	// the value of of the Managed App name so we must enforce a max of 36 even though
	// Azure supports a max of 64 chars for the Managed App name
	if !regexp.MustCompile(`^[-\da-zA-Z]{3,36}$`).MatchString(value) {
		msg := "must be between 3 and 36 characters in length and contains only letters, numbers or hyphens"
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg,
			AttributePath: path,
		})
	}

	return diagnostics
}

// validateAzureTags validates tags that will be applied to an Azure resource.
// Adapted from the azurerm provider.
// https://github.com/terraform-providers/terraform-provider-azurerm/blob/b7299d0b8c6f3685db07586530a7f52216dd48e4/azurerm/internal/tags/validation.go#L8
func validateAzureTags(v interface{}, path cty.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	tags := v.(map[string]interface{})

	if len(tags) > 50 {
		msg := "a maximum of 50 tags can be applied to each ARM resource"
		diagnostics = append(diagnostics, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       msg,
			Detail:        msg,
			AttributePath: path,
		})
	}

	for k, v := range tags {
		if len(k) > 512 {
			msg := fmt.Sprintf("the maximum length for a tag key is 512 characters: %q is %d characters", k, len(k))
			diagnostics = append(diagnostics, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       msg,
				Detail:        msg,
				AttributePath: path,
			})
		}

		value, err := helper.TagValueToString(v)
		if err != nil {
			diagnostics = append(diagnostics, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       err.Error(),
				Detail:        err.Error(),
				AttributePath: path,
			})
		} else if len(value) > 256 {
			msg := fmt.Sprintf("the maximum length for a tag value is 256 characters: the value for %q is %d characters", k, len(value))
			diagnostics = append(diagnostics, diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       msg,
				Detail:        msg,
				AttributePath: path,
			})
		}
	}

	return diagnostics
}
