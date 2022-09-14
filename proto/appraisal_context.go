// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package proto

func NewAppraisalContext(ec *EvidenceContext) *AppraisalContext {
	return &AppraisalContext{
		Evidence: ec,
		Result:   NewAttestationResult(ec),
	}
}
