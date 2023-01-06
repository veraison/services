// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package appraisal

import (
	"github.com/veraison/ear"
	"github.com/veraison/services/proto"
)

// Appraisal provides an appraisal context internally within the VTS (e.g. for
// policy evaluation). It is the analog of proto.AppraisalContext, but with a
// deserialized AttestationResult.
type Appraisal struct {
	EvidenceContext *proto.EvidenceContext
	Result          *ear.AttestationResult
	SignedEAR       []byte
}

func New(tenantID string) *Appraisal {
	return &Appraisal{
		EvidenceContext: &proto.EvidenceContext{
			TenantId: tenantID,
		},
		Result: ear.NewAttestationResult(),
	}
}

func (o *Appraisal) SetError() {
	if o.Result == nil {
		return
	}

	*o.Result.Status = ear.TrustTierNone
	o.Result.TrustVector.SetAll(ear.VerifierMalfunctionClaim)
}

func (o Appraisal) GetContext() *proto.AppraisalContext {
	return &proto.AppraisalContext{
		Evidence: o.EvidenceContext,
		Result:   o.SignedEAR,
	}
}
