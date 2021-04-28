// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// HashicorpCloudConsulamaAmaGetFederationRequest GetFederationRequest gets a complete federation view for the sole
// cluster found using the subscription ID and managed resource group. This can
// target a primary or secondary clusters.
//
// swagger:model hashicorp.cloud.consulama.ama.GetFederationRequest
type HashicorpCloudConsulamaAmaGetFederationRequest struct {

	// resource group
	ResourceGroup string `json:"resourceGroup,omitempty"`

	// subscription Id
	SubscriptionID string `json:"subscriptionId,omitempty"`
}

// Validate validates this hashicorp cloud consulama ama get federation request
func (m *HashicorpCloudConsulamaAmaGetFederationRequest) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *HashicorpCloudConsulamaAmaGetFederationRequest) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *HashicorpCloudConsulamaAmaGetFederationRequest) UnmarshalBinary(b []byte) error {
	var res HashicorpCloudConsulamaAmaGetFederationRequest
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
