// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// HashicorpCloudConsulamaAmaAuditLoggingUpdate AuditLoggingUpdate contains the details required to configure Consul audit logging
// for the cluster.
//
// swagger:model hashicorp.cloud.consulama.ama.AuditLoggingUpdate
type HashicorpCloudConsulamaAmaAuditLoggingUpdate struct {

	// enabled should be set to TRUE to enable audit logging.
	Enabled HashicorpCloudConsulamaAmaBoolean `json:"enabled,omitempty"`

	// storage_container_url is the URL of an Azure blob container to write
	// the Consul audit logs to. The vmss_identity must have write permissions
	// for this storage container.
	StorageContainerURL string `json:"storageContainerUrl,omitempty"`
}

// Validate validates this hashicorp cloud consulama ama audit logging update
func (m *HashicorpCloudConsulamaAmaAuditLoggingUpdate) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateEnabled(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *HashicorpCloudConsulamaAmaAuditLoggingUpdate) validateEnabled(formats strfmt.Registry) error {

	if swag.IsZero(m.Enabled) { // not required
		return nil
	}

	if err := m.Enabled.Validate(formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("enabled")
		}
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *HashicorpCloudConsulamaAmaAuditLoggingUpdate) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *HashicorpCloudConsulamaAmaAuditLoggingUpdate) UnmarshalBinary(b []byte) error {
	var res HashicorpCloudConsulamaAmaAuditLoggingUpdate
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
