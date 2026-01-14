// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package veraisonclient

import (
	"errors"

	"github.com/veraison/ear"
)

type ComponentVerifierClientHandler struct{}

func (s ComponentVerifierClientHandler) GetName() string {
	return "veraison-client-handler"
}

func (s ComponentVerifierClientHandler) GetAttestationScheme() string {
	return SchemeName
}

func (s ComponentVerifierClientHandler) GetSupportedMediaTypes() []string {
	return VeraisonClientMediaTypes
}

func (s ComponentVerifierClientHandler) AppraiseComponentEvidence(evidence []byte, mediaType string, nonce []byte) (*ear.Appraisal, error) {
	// TODO
	return nil, errors.New("not implemented")
}
