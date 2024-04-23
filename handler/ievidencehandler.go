// Copyright 2021-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"github.com/veraison/ear"
	"github.com/veraison/services/plugin"
	"github.com/veraison/services/proto"
)

// IEvidenceHandler defines the interface to functionality for working with
// attestation scheme specific evidence tokens. This includes validating token
// integrity, extracting and appraising claims.
type IEvidenceHandler interface {
	plugin.IPluggable

	// ExtractClaims parses the attestation token and returns claims
	// extracted therefrom.
	ExtractClaims(
		token *proto.AttestationToken,
		trustAnchors []string,
	) (map[string]interface{}, error)

	// ValidateEvidenceIntegrity verifies the structural integrity and validity of the
	// token. The exact checks performed are scheme-specific, but they
	// would typically involve, at the least, verifying the token's
	// signature using the provided trust anchors and endorsements. If the
	// validation fails, an error detailing what went wrong is returned.
	// Note: key material required to  validate the token would typically be
	//       provisioned as a Trust Anchor. However, depending on the
	//       requirements of the Scheme, it maybe be provisioned as an
	//       Endorsement instead, or in addition to the Trust Anchor. E.g.,
	//       if the validation is performed via an x.509 cert chain, the
	//       root cert may be provisioned as a Trust Anchor, while
	//       intermediate certs may be provisioned as Endorsements (at a
	//       different point in time, by a different actor).
	// TODO(setrofim): no distinction is currently made between validation
	// failing due to an internal error, and it failing due to bad input
	// (i.e. signature not matching).
	ValidateEvidenceIntegrity(
		token *proto.AttestationToken,
		trustAnchors []string,
		endorsementsStrings []string,
	) error

	// AppraiseEvidence evaluates the specified EvidenceContext against
	// the specified endorsements, and returns an AttestationResult.
	AppraiseEvidence(
		ec *proto.EvidenceContext,
		endorsements []string,
	) (*ear.AttestationResult, error)
}
