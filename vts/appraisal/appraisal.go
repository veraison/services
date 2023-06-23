// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package appraisal

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/veraison/ear"
	"github.com/veraison/services/config"
	"github.com/veraison/services/policy"
	"github.com/veraison/services/proto"
)

// Appraisal provides an appraisal context internally within the VTS (e.g. for
// policy evaluation). It is the analog of proto.AppraisalContext, but with a
// deserialized AttestationResult.
type Appraisal struct {
	Scheme          string
	EvidenceContext *proto.EvidenceContext
	Result          *ear.AttestationResult
	SignedEAR       []byte
}

func New(tenantID string, nonce []byte, scheme string) *Appraisal {
	appraisal := Appraisal{
		Scheme: scheme,
		EvidenceContext: &proto.EvidenceContext{
			TenantId: tenantID,
		},
		Result: ear.NewAttestationResult(scheme, config.Version, config.Developer),
	}

	encodedNonce := base64.URLEncoding.EncodeToString(nonce)
	appraisal.Result.Nonce = &encodedNonce

	appraisal.Result.VerifierID.Build = &config.Version
	appraisal.Result.VerifierID.Developer = &config.Developer

	appraisal.InitPolicyID()

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

func (o *Appraisal) UpdatePolicyID(pol *policy.Policy) error {
	if err := pol.Validate(); err != nil {
		return err
	}

	subID := pol.VersionedName()

	for _, submod := range o.Result.Submods {
		updatedID := strings.Join([]string{*submod.AppraisalPolicyID, subID}, "/")
		submod.AppraisalPolicyID = &updatedID
	}

	return nil
}

func (o *Appraisal) InitPolicyID() {
	for _, submod := range o.Result.Submods {
		policyID := fmt.Sprintf("policy:%s", o.Scheme)
		submod.AppraisalPolicyID = &policyID
	}
}
