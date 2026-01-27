// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package appraisal

import "github.com/veraison/services/proto"

// Evidence tracks the context of the evidence submmited for attestation.
type Evidence struct {
	TenantID  string `json:"tenant-id"`
	Data      []byte `json:"data"`
	MediaType string `json:"media-type"`
	Nonce     []byte `json:"nonce"`
}

// NewEvidenceFromProtobuf creates a new Evidence from a proto.AttestationToken
func NewEvidenceFromProtobuf(token *proto.AttestationToken) *Evidence {
	return &Evidence{
		TenantID:  token.TenantId,
		Data:      token.Data,
		MediaType: token.MediaType,
		Nonce:     token.Nonce,
	}
}

// ToProtobuf converts this Evidence to an proto.AttestationToken
func (o *Evidence) ToProtobuf() *proto.AttestationToken {
	return &proto.AttestationToken{
		TenantId:  o.TenantID,
		Data:      o.Data,
		MediaType: o.MediaType,
		Nonce:     o.Nonce,
	}
}
