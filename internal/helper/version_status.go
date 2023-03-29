// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package helper

import "github.com/hashicorp/terraform-provider-hcs/internal/clients/hcs-ama-api-spec/models"

// AMAVersionStatusToString converts a HashicorpCloudConsulamaAmaVersionStatus to a string.
func AMAVersionStatusToString(amaVersionStatus models.HashicorpCloudConsulamaAmaVersionStatus) string {
	var status string

	switch amaVersionStatus {
	case models.HashicorpCloudConsulamaAmaVersionStatusAVAILABLE:
		status = "AVAILABLE"
	case models.HashicorpCloudConsulamaAmaVersionStatusRECOMMENDED:
		status = "RECOMMENDED"
	case models.HashicorpCloudConsulamaAmaVersionStatusPREVIEW:
		status = "PREVIEW"
	}

	return status
}
