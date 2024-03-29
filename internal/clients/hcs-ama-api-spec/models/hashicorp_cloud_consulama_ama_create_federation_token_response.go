// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// HashicorpCloudConsulamaAmaCreateFederationTokenResponse CreateFederationTokenResponse contains the new Consul Federation Token.
//
// swagger:model hashicorp.cloud.consulama.ama.CreateFederationTokenResponse
type HashicorpCloudConsulamaAmaCreateFederationTokenResponse struct {

	// federation token
	FederationToken string `json:"federationToken,omitempty"`
}

// Validate validates this hashicorp cloud consulama ama create federation token response
func (m *HashicorpCloudConsulamaAmaCreateFederationTokenResponse) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *HashicorpCloudConsulamaAmaCreateFederationTokenResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *HashicorpCloudConsulamaAmaCreateFederationTokenResponse) UnmarshalBinary(b []byte) error {
	var res HashicorpCloudConsulamaAmaCreateFederationTokenResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
