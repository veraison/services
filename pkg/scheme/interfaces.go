// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package scheme

import (
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/ear"
	"github.com/veraison/services/pkg/plugin"
)

// IHandler defines the full interface that needs to be implemented by an
// attestation scheme. This includes the common pluggable interface used for
// runtime discovery.
type IHandler interface {
	plugin.IPluggable
	Implementation

	// GetAttestationScheme returns a string containing the name of the
	// attestation scheme handled by this IPluggable implementation.
	GetAttestationScheme() string

	// GetSupportedEvidenceMediaTypes returns the list of evidence media
	// types supported by this handler.
	GetSupportedEvidenceMediaTypes() []string

	// GetSupportedCorimProfiles returns the list of CoRIM profiles
	// supported by this handler.
	GetSupportedCorimProfiles() []string

	// ValidateCorim validates the contents of the provided CoRIM with respect
	// to the specified profile.
	ValidateCorim(uc *corim.UnsignedCorim) (*ValidateCorimResponse, error)

	// GetTrustAnchorIDs returns a slice of Environments used to
	// retrieve the trust anchors associated with evidence. The trust
	// anchors may be necessary to validate the entire evidence and/or extract
	// its claims (if it is encrypted).
	GetTrustAnchorIDs(evidence *Evidence) ([]*comid.Environment, error)

	// GetReferenceValueIDs returns a slice of Environments used to retrieve
	// reference values for an attestation scheme, using the claims
	// extracted from attestation token and the associated trust anchors.
	GetReferenceValueIDs(
		trustAnchors []*comid.KeyTriple,
		claims map[string]any,
	) ([]*comid.Environment, error)

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
	ValidateEvidenceIntegrity(
		evidence *Evidence,
		trustAnchors []*comid.KeyTriple,
		endorsements []*comid.ValueTriple,
	) error
}

// Implementation is the subset of the ISchemeHandler interface that
// must be implemented by schemes that make use of SchemeHandlerWrapper.
type Implementation interface {
	// ExtractClaims parses the attestation evidence and returns claims
	// extracted therefrom. (Note: the integrity of the input evidence must
	// be validated before any claims can be used; this can be done as part
	// of the extraction  process, or may be done as a separate stage by
	// implementing ValidateEvidenceIntegrity method defined in ISchemeHandler).
	ExtractClaims(
		evidence *Evidence,
		trustAnchors []*comid.KeyTriple,
	) (map[string]any, error)

	// AppraiseClaims evaluates the specified claims against
	// the specified endorsements, and returns an AttestationResult.
	AppraiseClaims(
		claims map[string]any,
		endorsements []*comid.ValueTriple,
	) (*ear.AttestationResult, error)

	// Optionally, if ValidateComid is implemented, without implementing ValidateCorim,
	// then ValidateComid will be invoked for every CoMID inside provisioned
	// each provisioned CoRIM.
	// ValidateComid(c *comid.Comid) error
}
