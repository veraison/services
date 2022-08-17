// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/services/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func Test_DecodeAttestationData_ok(t *testing.T) {
	data, err := os.ReadFile("test/tokens/basic.token")
	require.NoError(t, err)

	var decoded Token

	err = decoded.Decode(data)
	require.NoError(t, err)

	assert.Equal(t, uint32(4283712327), decoded.AttestationData.Magic)
	assert.Equal(t, uint64(0x7), decoded.AttestationData.FirmwareVersion)
}

func Test_GetTrustAnchorID_ok(t *testing.T) {
	data, err := os.ReadFile("test/tokens/basic.token")
	require.NoError(t, err)

	ta := proto.AttestationToken{
		TenantId: "0",
		Format:   proto.AttestationFormat_TPM_ENACTTRUST,
		Data:     data,
	}

	var s Scheme

	taID, err := s.GetTrustAnchorID(&ta)
	require.NoError(t, err)
	assert.Equal(t, "TPM_ENACTTRUST://0/7df7714e-aa04-4638-bcbf-434b1dd720f1", taID)
}

func readPublicKeyBytes(path string) ([]byte, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(buf)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("could not decode EC public key from PEM block: %q", block)
	}
	return block.Bytes, nil
}

func Test_ExtractVerifiedClaims_ok(t *testing.T) {
	data, err := os.ReadFile("test/tokens/basic.token")
	require.NoError(t, err)

	ta := proto.AttestationToken{
		TenantId: "0",
		Format:   proto.AttestationFormat_TPM_ENACTTRUST,
		Data:     data,
	}

	var s Scheme

	trustAnchorBytes, err := readPublicKeyBytes("test/keys/basic.pem.pub")
	require.NoError(t, err)
	trustAnchor := base64.StdEncoding.EncodeToString(trustAnchorBytes)

	ev, err := s.ExtractVerifiedClaims(&ta, trustAnchor)
	require.Nil(t, err)

	expectedPCRDigest := []byte{
		0x87, 0x42, 0x8f, 0xc5, 0x22, 0x80, 0x3d, 0x31, 0x6, 0x5e, 0x7b, 0xce,
		0x3c, 0xf0, 0x3f, 0xe4, 0x75, 0x9, 0x66, 0x31, 0xe5, 0xe0, 0x7b, 0xbd,
		0x7a, 0xf, 0xde, 0x60, 0xc4, 0xcf, 0x25, 0xc7,
	}

	assert.Equal(t, "TPM_ENACTTRUST://0/7df7714e-aa04-4638-bcbf-434b1dd720f1", ev.SoftwareID)
	assert.Equal(t, []int64{1, 2, 3, 4}, ev.ClaimsSet["pcr-selection"])
	assert.Equal(t, int64(11), ev.ClaimsSet["hash-algorithm"])
	assert.Equal(t, expectedPCRDigest, ev.ClaimsSet["pcr-digest"])
}

func Test_GetAttestation(t *testing.T) {
	evStruct, err := structpb.NewStruct(map[string]interface{}{
		"pcr-selection":  []interface{}{1, 2, 3, 4},
		"hash-algorithm": int64(11),
		"pcr-digest": []byte{
			0x87, 0x42, 0x8f, 0xc5, 0x22, 0x80, 0x3d, 0x31, 0x6, 0x5e, 0x7b,
			0xce, 0x3c, 0xf0, 0x3f, 0xe4, 0x75, 0x9, 0x66, 0x31, 0xe5, 0xe0,
			0x7b, 0xbd, 0x7a, 0xf, 0xde, 0x60, 0xc4, 0xcf, 0x25, 0xc7,
		},
	})
	require.NoError(t, err)

	evidenceContext := &proto.EvidenceContext{
		Format:        proto.AttestationFormat_TPM_ENACTTRUST,
		TenantId:      "0",
		TrustAnchorId: "TPM_ENACTTRUST://0/7df7714e-aa04-4638-bcbf-434b1dd720f1",
		SoftwareId:    "TPM_ENACTTRUST://0/7df7714e-aa04-4638-bcbf-434b1dd720f1",
		Evidence:      evStruct,
	}
	endorsements := []string{"h0KPxSKAPTEGXnvOPPA/5HUJZjHl4Hu9eg/eYMTPJcc="}

	var scheme Scheme

	appraisalCtx, err := scheme.AppraiseEvidence(evidenceContext, endorsements)
	require.NoError(t, err)

	assert.Equal(t, proto.AR_Status_SUCCESS, appraisalCtx.Result.Status)
	assert.Equal(t, proto.AR_Status_SUCCESS, appraisalCtx.Result.TrustVector.SoftwareIntegrity)
}
