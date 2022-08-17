// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package scheme

import "github.com/veraison/services/proto"

// IScheme defines the interface to attestation scheme specific functionality.
// An object implementing this interface encapsulates all functionality specific
// to a particular AttestationFormat, such as knowledge of evidence and
// endorsements structure.
type IScheme interface {
	GetName() string
	GetFormat() proto.AttestationFormat
	GetSupportedMediaTypes() []string

	ExtractVerifiedClaims(token *proto.AttestationToken, trustAnchor string) (*ExtractedClaims, error)
	GetTrustAnchorID(token *proto.AttestationToken) (string, error)
	AppraiseEvidence(ec *proto.EvidenceContext, endorsements []string) (*proto.AppraisalContext, error)

	// endorsement lookup keys
	SynthKeysFromSwComponent(tenantID string, swComp *proto.Endorsement) ([]string, error)
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
