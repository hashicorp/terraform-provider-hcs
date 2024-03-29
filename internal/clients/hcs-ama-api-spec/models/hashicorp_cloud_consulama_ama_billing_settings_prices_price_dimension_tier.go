// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// HashicorpCloudConsulamaAmaBillingSettingsPricesPriceDimensionTier Tier represents one billing tier of a price dimension.
//
// swagger:model hashicorp.cloud.consulama.ama.BillingSettings.Prices.PriceDimension.Tier
type HashicorpCloudConsulamaAmaBillingSettingsPricesPriceDimensionTier struct {

	// label is a string description of this pricing tier.
	Label string `json:"label,omitempty"`

	// unit_price is the price per unit billed based on this tier.
	UnitPrice float64 `json:"unitPrice,omitempty"`
}

// Validate validates this hashicorp cloud consulama ama billing settings prices price dimension tier
func (m *HashicorpCloudConsulamaAmaBillingSettingsPricesPriceDimensionTier) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *HashicorpCloudConsulamaAmaBillingSettingsPricesPriceDimensionTier) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *HashicorpCloudConsulamaAmaBillingSettingsPricesPriceDimensionTier) UnmarshalBinary(b []byte) error {
	var res HashicorpCloudConsulamaAmaBillingSettingsPricesPriceDimensionTier
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
