// Copyright 2021-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package scheme

import (
	"github.com/veraison/ear"
	"github.com/veraison/services/proto"
)

// IScheme defines the interface to attestation scheme specific functionality.
// An object implementing this interface encapsulates all functionality specific
// to a particular AttestationFormat, such as knowledge of evidence and
// endorsements structure.
type IScheme interface {
	// GetName returns the Scheme name as a string. The name is
	// typically taken from the corresponding attestation format (it is,
	// however, theoretically possible for them to be different, e.g. if
	// there are multiple schemes associated with the same format) .
	GetName() string

	// GetSupportedMediaTypes returns a string slice of MIME media types
	// handled by this scheme.
	GetSupportedMediaTypes() []string

	// GetTrustAnchorID returns a string ID used to retrieve a trust anchor
	// for this token. The trust anchor may be necessary to validate the
	// token and/or extract its claims (if it is encrypted).
	GetTrustAnchorID(token *proto.AttestationToken) (string, error)

	// ExtractClaims parses the attestation token and returns claims
	// extracted therefrom.
	ExtractClaims(
		token *proto.AttestationToken,
		trustAnchor string,
	) (*ExtractedClaims, error)

	// ValidateEvidenceIntegrity verifies the structural integrity and validity of the
	// token. The exact checks performed are scheme-specific, but they
	// would typically involve, at the least, verifying the token's
	// signature using the provided trust anchor and endorsements. If the
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
		trustAnchor string,
		endorsementsStrings []string,
	) error

	// AppraiseEvidence evaluates the specified  EvidenceContext against
	// the specified endorsements, and returns an AttestationResult.
	AppraiseEvidence(
		ec *proto.EvidenceContext,
		endorsements []string,
	) (*ear.AttestationResult, error)

	// SynthKeysFromSwComponent synthesizes lookup key(s) for the
	// provided software component endorsement.
	SynthKeysFromSwComponent(tenantID string, swComp *proto.Endorsement) ([]string, error)

	// SynthKeysFromTrustAnchor synthesizes lookup key(s) for the provided
	// trust anchor.
	SynthKeysFromTrustAnchor(tenantID string, ta *proto.Endorsement) ([]string, error)
}

// ExtractedClaims contains a map of claims extracted from an attestation
// token along with the corresponding SoftwareID that is used to fetch
// the associated endorsements.
//
// XXX(tho) -- not clear why SoftwareID is treated differently from TrustAnchorID
type ExtractedClaims struct {
	ClaimsSet  map[string]interface{} `json:"claims-set"`
	SoftwareID string                 `json:"software-id"`
}

func NewExtractedClaims() *ExtractedClaims {
	return &ExtractedClaims{
		ClaimsSet: make(map[string]interface{}),
	}
}
