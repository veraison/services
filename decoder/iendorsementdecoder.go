// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package decoder

import (
	"github.com/veraison/services/plugin"
)

// EndorsementDecoderParams are passed to IEvidenceDecoder.Init() They are
// implementation-specific.
type EndorsementDecoderParams map[string]interface{}

// IEndorsementDecoder defines the interface to functionality for working with
// attestation scheme specific endorsement provisioning tokens (typically,
// CoRIM's).
type IEndorsementDecoder interface {
	plugin.IPluggable

	// Init() initializes the decoder.
	Init(params EndorsementDecoderParams) error

	// Close the decoder, finalizing any state it may contain.
	Close() error

	// Decoder the endorsements from the provided []byte.
	Decode([]byte) (*EndorsementDecoderResponse, error)
}
