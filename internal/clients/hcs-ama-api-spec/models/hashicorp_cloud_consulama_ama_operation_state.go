// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

// HashicorpCloudConsulamaAmaOperationState State is one of the states that an Operation can be in.
//
// The states are purposely coarse grained to make it easy to understand the operation
// state machine: pending => running => done. No other state transitions are possible.
// Success/failure can be determined based on the result oneof.
//
// swagger:model hashicorp.cloud.consulama.ama.Operation.State
type HashicorpCloudConsulamaAmaOperationState string

const (

	// HashicorpCloudConsulamaAmaOperationStatePENDING captures enum value "PENDING"
	HashicorpCloudConsulamaAmaOperationStatePENDING HashicorpCloudConsulamaAmaOperationState = "PENDING"

	// HashicorpCloudConsulamaAmaOperationStateRUNNING captures enum value "RUNNING"
	HashicorpCloudConsulamaAmaOperationStateRUNNING HashicorpCloudConsulamaAmaOperationState = "RUNNING"

	// HashicorpCloudConsulamaAmaOperationStateDONE captures enum value "DONE"
	HashicorpCloudConsulamaAmaOperationStateDONE HashicorpCloudConsulamaAmaOperationState = "DONE"
)

// for schema
var hashicorpCloudConsulamaAmaOperationStateEnum []interface{}

func init() {
	var res []HashicorpCloudConsulamaAmaOperationState
	if err := json.Unmarshal([]byte(`["PENDING","RUNNING","DONE"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		hashicorpCloudConsulamaAmaOperationStateEnum = append(hashicorpCloudConsulamaAmaOperationStateEnum, v)
	}
}

func (m HashicorpCloudConsulamaAmaOperationState) validateHashicorpCloudConsulamaAmaOperationStateEnum(path, location string, value HashicorpCloudConsulamaAmaOperationState) error {
	if err := validate.EnumCase(path, location, value, hashicorpCloudConsulamaAmaOperationStateEnum, true); err != nil {
		return err
	}
	return nil
}

// Validate validates this hashicorp cloud consulama ama operation state
func (m HashicorpCloudConsulamaAmaOperationState) Validate(formats strfmt.Registry) error {
	var res []error

	// value enum
	if err := m.validateHashicorpCloudConsulamaAmaOperationStateEnum("", "body", m); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
