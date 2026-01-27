// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/services/plugin"
	"github.com/veraison/services/vts/appraisal"
)

// ISchemeHandler defines the full interface that needs to be implemented by an
// attestation scheme. This includes the common pluggable interface used for
// runtime discovery.
type ISchemeHandler interface {
	plugin.IPluggable
	ISchemeImplementation

	// GetSupportedProvisioningMediaTypes returns the list of media types
	// supported on the provisioning path; i.e. the supported endorsement
	// formats.
	GetSupportedProvisioningMediaTypes() []string

	// GetSupportedVerificationMediaTypes returns the list of media types
	// supported on the verification path; i.e. the supported attestation
	// evidence formats.
	GetSupportedVerificationMediaTypes() []string

	// ValidateCorim validates the contents of the provided CoRIM with respect
	// to the specified profile.
	ValidateCorim(uc *corim.UnsignedCorim) (*ValidateCorimResponse, error)

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
		evidence *appraisal.Evidence,
		trustAnchors []*comid.KeyTriple,
		endorsements []*comid.ValueTriple,
	) error
}
