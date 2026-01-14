// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"github.com/veraison/ear"
	"github.com/veraison/services/plugin"
)

// IComponentVerifierClientHandler defines the interface for component verifier clients
type IComponentVerifierClientHandler interface {
	plugin.IPluggable

	// TODO(tho): add a configuration parameter
	AppraiseComponentEvidence(evidence []byte, mediaType string, nonce []byte) (*ear.Appraisal, error)
}
