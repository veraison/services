// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package appraisal

import (
	"encoding/base64"

	"github.com/veraison/ear"
	"github.com/veraison/services/config"
	"github.com/veraison/services/proto"
)

// Appraisal provides an appraisal context internally within the VTS (e.g. for
// policy evaluation). It is the analog of proto.AppraisalContext, but with a
// deserialized AttestationResult.
type Appraisal struct {
	EvidenceContext *proto.EvidenceContext
	Result          *ear.AttestationResult
	SignedEAR       []byte
	TeeReport       bool
}

func New(tenantID string, nonce []byte, submodName string, teeReport bool) *Appraisal {
	appraisal := Appraisal{
		EvidenceContext: &proto.EvidenceContext{
			TenantId: tenantID,
		},
		Result:    ear.NewAttestationResult(submodName, config.Version, config.Developer),
		TeeReport: teeReport,
	}

	encodedNonce := base64.URLEncoding.EncodeToString(nonce)
	appraisal.Result.Nonce = &encodedNonce

	appraisal.Result.VerifierID.Build = &config.Version
	appraisal.Result.VerifierID.Developer = &config.Developer

	return &appraisal
}

func (o Appraisal) GetContext() *proto.AppraisalContext {
	return &proto.AppraisalContext{
		Evidence: o.EvidenceContext,
		Result:   o.SignedEAR,
	}
}

func (o Appraisal) SetAllClaims(claim ear.TrustClaim) {
	for _, submod := range o.Result.Submods {
		submod.TrustVector.SetAll(claim)
	}
}

func (o Appraisal) AddPolicyClaim(name, claim string) {
	for _, submod := range o.Result.Submods {
		if submod.AppraisalExtensions.VeraisonPolicyClaims == nil {
			claimsMap := make(map[string]interface{})
			submod.AppraisalExtensions.VeraisonPolicyClaims = &claimsMap
		}
		(*submod.AppraisalExtensions.VeraisonPolicyClaims)[name] = claim
	}
}
