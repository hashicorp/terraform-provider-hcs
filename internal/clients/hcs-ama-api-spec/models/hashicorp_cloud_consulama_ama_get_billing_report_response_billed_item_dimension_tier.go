// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// HashicorpCloudConsulamaAmaGetBillingReportResponseBilledItemDimensionTier Tier represents the usage that is billed according to one tier.
//
// swagger:model hashicorp.cloud.consulama.ama.GetBillingReportResponse.BilledItem.Dimension.Tier
type HashicorpCloudConsulamaAmaGetBillingReportResponseBilledItemDimensionTier struct {

	// label is a string description of this tier.
	Label string `json:"label,omitempty"`

	// units is the number of consumed units in this tier.
	Units int32 `json:"units,omitempty"`
}

// Validate validates this hashicorp cloud consulama ama get billing report response billed item dimension tier
func (m *HashicorpCloudConsulamaAmaGetBillingReportResponseBilledItemDimensionTier) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *HashicorpCloudConsulamaAmaGetBillingReportResponseBilledItemDimensionTier) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *HashicorpCloudConsulamaAmaGetBillingReportResponseBilledItemDimensionTier) UnmarshalBinary(b []byte) error {
	var res HashicorpCloudConsulamaAmaGetBillingReportResponseBilledItemDimensionTier
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
