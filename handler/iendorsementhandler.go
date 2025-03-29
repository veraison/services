// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"github.com/veraison/services/plugin"
)

// EndorsementHandlerParams are passed to IEndorsementHandler.Init() They are
// implementation-specific.
type EndorsementHandlerParams map[string]interface{}

// IEndorsementHandler defines the interface to functionality for working with
// attestation scheme specific endorsement provisioning tokens (typically,
// CoRIM's).
type IEndorsementHandler interface {
	plugin.IPluggable

	// Init() initializes the handler.
	Init(params EndorsementHandlerParams) error

	// Close the decoder, finalizing any state it may contain.
	Close() error

	// Decode the endorsements from the provided []byte. The mediaType parameter
	// specifies the format of the input data (e.g. "application/rim+cbor" for
	// unsigned CoRIMs or "application/rim+cose" for signed CoRIMs).
	Decode([]byte, string) (*EndorsementHandlerResponse, error)
}
