// Copyright 2022-2026 Contributors to the Veraison project.
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
	"google.golang.org/protobuf/types/known/structpb"
)

// Appraisal provides an appraisal context internally within the VTS (e.g. for
// policy evaluation).
type Appraisal struct {
	Scheme          string
	EvidenceContext *proto.EvidenceContext
	Result          *ear.AttestationResult
	SignedEAR       []byte
	Endorsements    []string
}

func New(tenantID string, nonce []byte, scheme string) *Appraisal {
	appraisal := Appraisal{
		Scheme: scheme,
		EvidenceContext: &proto.EvidenceContext{
			TenantId: tenantID,
		},
		Result: ear.NewAttestationResult(scheme, config.Version, config.Developer),
	}

	appraisal.setResultNonce(nonce)
	appraisal.InitPolicyID()

	return &appraisal
}

func (o *Appraisal) setResultNonce(v []byte) {
	encodedNonce := base64.URLEncoding.EncodeToString(v)
	o.Result.Nonce = &encodedNonce
}

func (o *Appraisal) SetTrustAnchorIDs(v []string) {
	o.EvidenceContext.TrustAnchorIds = v
}

func (o *Appraisal) SetReferenceIDs(v []string) {
	o.EvidenceContext.ReferenceIds = v
}

func (o Appraisal) GetReferenceIDs() []string {
	return o.EvidenceContext.ReferenceIds
}

func (o *Appraisal) SetEvidenceClaims(v *structpb.Struct) {
	o.EvidenceContext.Evidence = v
}

func (o *Appraisal) SetEndorsements(v []string) {
	o.Endorsements = v
}

func (o *Appraisal) SetResultWithNonce(result *ear.AttestationResult, nonce []byte) {
	o.Result = result
	o.setResultNonce(nonce)
}

func (o Appraisal) GetContext() *proto.AppraisalContext {
	return &proto.AppraisalContext{
		Evidence: o.EvidenceContext,
		Result:   o.SignedEAR,
	}
}

func (o *Appraisal) SetAllClaims(claim ear.TrustClaim) {
	for _, submod := range o.Result.Submods {
		submod.TrustVector.SetAll(claim)
	}
}

func (o *Appraisal) AddPolicyClaim(name, claim string) {
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

	subID := pol.UUID.String()

	for _, submod := range o.Result.Submods {
		updatedID := strings.Join([]string{*submod.AppraisalPolicyID, subID}, "/")
		submod.AppraisalPolicyID = &updatedID
	}

	return nil
}

// InitPolicyID must be called before sending the appraisal to policy manager
// for evaluation.
func (o *Appraisal) InitPolicyID() {
	for _, submod := range o.Result.Submods {
		policyID := fmt.Sprintf("policy:%s", o.Scheme)
		submod.AppraisalPolicyID = &policyID
	}
}
