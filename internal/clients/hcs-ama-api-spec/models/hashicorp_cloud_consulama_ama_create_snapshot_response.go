// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// HashicorpCloudConsulamaAmaCreateSnapshotResponse See ConsulAMAService.CreateSnapshot
//
// swagger:model hashicorp.cloud.consulama.ama.CreateSnapshotResponse
type HashicorpCloudConsulamaAmaCreateSnapshotResponse struct {

	// operation used to track the progress of the snapshot create
	Operation *HashicorpCloudConsulamaAmaOperation `json:"operation,omitempty"`

	// snapshot_id is the ID of the Consul snapshot to create.
	SnapshotID string `json:"snapshotId,omitempty"`
}

// Validate validates this hashicorp cloud consulama ama create snapshot response
func (m *HashicorpCloudConsulamaAmaCreateSnapshotResponse) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateOperation(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *HashicorpCloudConsulamaAmaCreateSnapshotResponse) validateOperation(formats strfmt.Registry) error {

	if swag.IsZero(m.Operation) { // not required
		return nil
	}

	if m.Operation != nil {
		if err := m.Operation.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("operation")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *HashicorpCloudConsulamaAmaCreateSnapshotResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *HashicorpCloudConsulamaAmaCreateSnapshotResponse) UnmarshalBinary(b []byte) error {
	var res HashicorpCloudConsulamaAmaCreateSnapshotResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
