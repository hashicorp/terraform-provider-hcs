// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"strconv"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
)

// HashicorpCloudConsulamaAmaListClustersResponse See ConsulAMAService.ListCluster
// swagger:model hashicorp.cloud.consulama.ama.ListClustersResponse
type HashicorpCloudConsulamaAmaListClustersResponse struct {

	// value is the list of cluster in the format of ClusterResponses.
	Value []*HashicorpCloudConsulamaAmaClusterResponse `json:"value"`
}

// Validate validates this hashicorp cloud consulama ama list clusters response
func (m *HashicorpCloudConsulamaAmaListClustersResponse) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateValue(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *HashicorpCloudConsulamaAmaListClustersResponse) validateValue(formats strfmt.Registry) error {

	if swag.IsZero(m.Value) { // not required
		return nil
	}

	for i := 0; i < len(m.Value); i++ {
		if swag.IsZero(m.Value[i]) { // not required
			continue
		}

		if m.Value[i] != nil {
			if err := m.Value[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("value" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *HashicorpCloudConsulamaAmaListClustersResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *HashicorpCloudConsulamaAmaListClustersResponse) UnmarshalBinary(b []byte) error {
	var res HashicorpCloudConsulamaAmaListClustersResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
