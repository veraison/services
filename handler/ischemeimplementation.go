// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"github.com/veraison/corim/comid"
	"github.com/veraison/ear"
	"github.com/veraison/services/vts/appraisal"
)

// ISchemeImplementation is the subset of the ISchemeHandler interface that
// must be implemented by schemes that make use of SchemeHandlerWrapper.
type ISchemeImplementation interface {
	// GetTrustAnchorIDs returns a slice of Environements used to
	// retrieve the trust anchors associated with evidence. The trust
	// anchors may be necessary to validate the entire evidence and/or extract
	// its claims (if it is encrypted).
	GetTrustAnchorIDs(evidence *appraisal.Evidence) ([]*comid.Environment, error)

	// ExtractClaims parses the attestation evidence and returns claims
	// extracted therefrom.
	ExtractClaims(
		evidence *appraisal.Evidence,
		trustAnchors []*comid.KeyTriple,
	) (map[string]any, error)

	// AppraiseClaims evaluates the specified claims against
	// the specified endorsements, and returns an AttestationResult.
	AppraiseClaims(
		claims map[string]any,
		endorsements []*comid.ValueTriple,
	) (*ear.AttestationResult, error)

	// Optionally, if ValidateComid is implemented, and ValidateCorim is not,
	// then ValidateComid will be invoked for every CoMID inside provisioned
	// each provisioned CoRIM.
	// ValidateComid(c *comid.Comid) error
}

