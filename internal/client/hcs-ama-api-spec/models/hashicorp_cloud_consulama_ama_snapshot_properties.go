// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// HashicorpCloudConsulamaAmaSnapshotProperties SnapshotProperties contains properties of the Consul snapshot
// swagger:model hashicorp.cloud.consulama.ama.SnapshotProperties
type HashicorpCloudConsulamaAmaSnapshotProperties struct {

	// finished_at notes the time that this snapshot was finished.
	// Format: date-time
	FinishedAt strfmt.DateTime `json:"finishedAt,omitempty"`

	// id is the id of the snapshot.
	ID string `json:"id,omitempty"`

	// name is the name of the snapshot.
	Name string `json:"name,omitempty"`

	// product_version is the version of the product of the cluster at creation.
	ProductVersion string `json:"productVersion,omitempty"`

	// requested_at notes the time that this snapshot was requested.
	// Format: date-time
	RequestedAt strfmt.DateTime `json:"requestedAt,omitempty"`

	// restored_at notes the time that this snapshot was restored.
	// Format: date-time
	RestoredAt strfmt.DateTime `json:"restoredAt,omitempty"`

	// size is the size of the snapshot in bytes.
	Size string `json:"size,omitempty"`

	// state is the current state of the snapshot.
	State string `json:"state,omitempty"`

	// type is the type of snapshot; MANUAL or AUTOMATIC
	Type string `json:"type,omitempty"`
}

// Validate validates this hashicorp cloud consulama ama snapshot properties
func (m *HashicorpCloudConsulamaAmaSnapshotProperties) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateFinishedAt(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateRequestedAt(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateRestoredAt(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *HashicorpCloudConsulamaAmaSnapshotProperties) validateFinishedAt(formats strfmt.Registry) error {

	if swag.IsZero(m.FinishedAt) { // not required
		return nil
	}

	if err := validate.FormatOf("finishedAt", "body", "date-time", m.FinishedAt.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *HashicorpCloudConsulamaAmaSnapshotProperties) validateRequestedAt(formats strfmt.Registry) error {

	if swag.IsZero(m.RequestedAt) { // not required
		return nil
	}

	if err := validate.FormatOf("requestedAt", "body", "date-time", m.RequestedAt.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *HashicorpCloudConsulamaAmaSnapshotProperties) validateRestoredAt(formats strfmt.Registry) error {

	if swag.IsZero(m.RestoredAt) { // not required
		return nil
	}

	if err := validate.FormatOf("restoredAt", "body", "date-time", m.RestoredAt.String(), formats); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *HashicorpCloudConsulamaAmaSnapshotProperties) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *HashicorpCloudConsulamaAmaSnapshotProperties) UnmarshalBinary(b []byte) error {
	var res HashicorpCloudConsulamaAmaSnapshotProperties
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
