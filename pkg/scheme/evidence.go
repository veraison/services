// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package scheme

// Evidence tracks the context of the evidence submitted for attestation.
type Evidence struct {
	Data      []byte `cbor:"1,keyasint" json:"data"`
	MediaType string `cbor:"2,keyasint" json:"media-type"`
	Nonce     []byte `cbor:"3,keyasint" json:"nonce"`
}
