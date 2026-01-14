// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"github.com/veraison/services/plugin"
)

// IComponentVerifierClientHandler defines the interface for component verifier clients
type IComponentVerifierClientHandler interface {
	plugin.IPluggable

	// AppraiseComponentEvidence forwards evidence (and nonce) to the component
	// verifier and returns the received attestation results as an JSON-encoded
	// EAR appraisal(s).  The clientCfg parameter is a client-specific
	// configuration supplied as a JSON-encoded byte array.
	AppraiseComponentEvidence(
		evidence []byte,
		mediaType string,
		nonce []byte,
		clientCfg []byte,
	) ([]byte, error)
}
