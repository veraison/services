// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package verifier

import (
	"github.com/veraison/services/proto"
)

type IVerifier interface {
	GetVTSState() (*proto.ServiceState, error)
	GetPublicKey() (*proto.PublicKey, error)
	IsSupportedMediaType(mt string) (bool, error)
	IsSupportedCompositeEvidenceMediaType(mt string) (bool, error)
	SupportedMediaTypes() ([]string, error)
	SupportedCompositeEvidenceMediaTypes() ([]string, error)
	ProcessEvidence(tenantID string, nonce []byte, data []byte, mt string) ([]byte, error)
	ProcessCompositeEvidence(tenantID string, nonce []byte, data []byte, mt string) ([]byte, error)
}
