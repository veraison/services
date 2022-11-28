// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	nitro_enclave_attestation_document "github.com/veracruz-project/go-nitro-enclave-attestation-document"
	"github.com/veraison/services/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testTime time.Time = time.Date(2022, 11, 9, 23, 0, 0, 0, time.UTC)

func generateValidTimeRange(expired bool) (time.Time, time.Time) {
	var notBefore time.Time
	var notAfter time.Time
	if expired {
		notBefore = time.Now().Add(-time.Hour * 24)
		notAfter = time.Now().Add(-time.Hour * 1)
	} else {
		notBefore = time.Now()
		notAfter = time.Now().Add(time.Hour * 24 * 180)
	}
	return notBefore, notAfter
}

func generateCertsAndKeys(endCertExpired bool, caCertExpired bool) (*ecdsa.PrivateKey, []byte, *x509.Certificate, []byte, error) {
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to generate CA key:%v", err)
	}

	caNotBefore, caNotAfter := generateValidTimeRange(caCertExpired)
	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: caNotBefore,
		NotAfter:  caNotAfter,

		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	caCertDer, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("Failed to generate CA Certificate:%v", err)
	}
	caCert, err := x509.ParseCertificate(caCertDer)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("Failed to convert CA Cert der to certificate:%v", err)
	}

	endKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("Failed to generate end key:%v", err)
	}

	endNotBefore, endNotAfter := generateValidTimeRange(endCertExpired)
	endTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: endNotBefore,
		NotAfter:  endNotAfter,

		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	endCertDer, err := x509.CreateCertificate(rand.Reader, &endTemplate, caCert, &endKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("Failed to generate end certificate:%v", err)
	}
	return endKey, endCertDer, caCert, caCertDer, nil
}

const NUM_PCRS = 16

func generateRandomSlice(size int32) []byte {
	result := make([]byte, size)
	rand.Read(result)
	return result
}

func generatePCRs() (map[int32][]byte, error) {
	pcrs := make(map[int32][]byte)
	for i := int32(0); i < NUM_PCRS; i++ {
		pcrs[i] = generateRandomSlice(96)
	}
	return pcrs, nil
}

func genTaEndorsements(caCertDer []byte) ([]byte, error) {
	taEndValBytes, err := os.ReadFile("test/ta-endorsements.json")
	if err != nil {
		return nil, fmt.Errorf("os.ReadFile failed:%v\n", err)
	}
	var pemCertBlock = &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertDer,
	}
	caCertPem := string(pem.EncodeToMemory(pemCertBlock))
	caCertJson, err := json.Marshal(caCertPem)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal failed:%v", err)
	}
	taEndValString := string(taEndValBytes)
	taEndValString = strings.Replace(taEndValString, "\"<CERT>\"", string(caCertJson), 1)
	taEndValBytes = []byte(taEndValString)
	return taEndValBytes, nil
}

func Test_GetTrustAnchorID_ok(t *testing.T) {
	privateKey, endCertDer, _, caCertDer, err := generateCertsAndKeys(false, false)
	require.NoError(t, err)

	PCRs, err := generatePCRs()
	require.NoError(t, err)
	userData := generateRandomSlice(32)
	nonce := generateRandomSlice(32)
	tokenBytes, err := nitro_enclave_attestation_document.GenerateDocument(PCRs, userData, nonce, endCertDer, [][]byte{caCertDer}, privateKey)
	require.NoError(t, err)

	token := proto.AttestationToken{
		TenantId: "1",
		Format:   proto.AttestationFormat_PSA_IOT,
		Data:     tokenBytes,
	}

	expectedTaID := "AWS_NITRO://1/"

	scheme := &Scheme{}

	taID, err := scheme.GetTrustAnchorID(&token)
	require.NoError(t, err)
	assert.Equal(t, expectedTaID, taID)
}

func Test_ExtractVerifiedClaims_ok(t *testing.T) {
	privateKey, endCertDer, _, caCertDer, err := generateCertsAndKeys(false, false)
	require.NoError(t, err)

	PCRs, err := generatePCRs()
	require.NoError(t, err)
	userData := generateRandomSlice(32)
	nonce := generateRandomSlice(32)
	tokenBytes, err := nitro_enclave_attestation_document.GenerateDocument(PCRs, userData, nonce, endCertDer, [][]byte{caCertDer}, privateKey)
	require.NoError(t, err)

	taEndValBytes, err := genTaEndorsements(caCertDer)
	require.NoError(t, err)

	scheme := &Scheme{}

	token := proto.AttestationToken{
		TenantId: "1",
		Format:   proto.AttestationFormat_AWS_NITRO,
		Data:     tokenBytes,
	}

	extracted, err := scheme.ExtractClaims(&token, string(taEndValBytes))

	require.NoError(t, err)
	assert.Equal(t, PCRs[0][:], extracted.ClaimsSet["PCR0"].([]byte))
	assert.Equal(t, PCRs[1][:], extracted.ClaimsSet["PCR1"].([]byte))
	assert.Equal(t, PCRs[2][:], extracted.ClaimsSet["PCR2"].([]byte))
	assert.Equal(t, PCRs[3][:], extracted.ClaimsSet["PCR3"].([]byte))
	assert.Equal(t, PCRs[4][:], extracted.ClaimsSet["PCR4"].([]byte))
	assert.Equal(t, PCRs[5][:], extracted.ClaimsSet["PCR5"].([]byte))
	assert.Equal(t, PCRs[6][:], extracted.ClaimsSet["PCR6"].([]byte))
	assert.Equal(t, PCRs[7][:], extracted.ClaimsSet["PCR7"].([]byte))
	assert.Equal(t, PCRs[8][:], extracted.ClaimsSet["PCR8"].([]byte))
	assert.Equal(t, PCRs[9][:], extracted.ClaimsSet["PCR9"].([]byte))
	assert.Equal(t, PCRs[10][:], extracted.ClaimsSet["PCR10"].([]byte))
	assert.Equal(t, PCRs[11][:], extracted.ClaimsSet["PCR11"].([]byte))
	assert.Equal(t, PCRs[12][:], extracted.ClaimsSet["PCR12"].([]byte))
	assert.Equal(t, PCRs[13][:], extracted.ClaimsSet["PCR13"].([]byte))
	assert.Equal(t, PCRs[14][:], extracted.ClaimsSet["PCR14"].([]byte))
	assert.Equal(t, PCRs[15][:], extracted.ClaimsSet["PCR15"].([]byte))

	receivedNonce := extracted.ClaimsSet["nonce"].([]byte)
	assert.Equal(t, nonce[:], receivedNonce[:])

	receivedUserData := extracted.ClaimsSet["user_data"].([]byte)
	assert.Equal(t, userData[:], receivedUserData[:])
}

func Test_ExtractVerifiedClaims_bad_signature(t *testing.T) {
	privateKey, endCertDer, _, caCertDer, err := generateCertsAndKeys(false, false)
	require.NoError(t, err)

	PCRs, err := generatePCRs()
	require.NoError(t, err)
	userData := generateRandomSlice(32)
	nonce := generateRandomSlice(32)
	tokenBytes, err := nitro_enclave_attestation_document.GenerateDocument(PCRs, userData, nonce, endCertDer, [][]byte{caCertDer}, privateKey)
	require.NoError(t, err)

	// modify the signature to make it fail
	tokenBytes[len(tokenBytes)-1] ^= tokenBytes[len(tokenBytes)-1]

	taEndValBytes, err := genTaEndorsements(caCertDer)
	require.NoError(t, err)

	scheme := &Scheme{}

	token := proto.AttestationToken{
		TenantId: "1",
		Format:   proto.AttestationFormat_AWS_NITRO,
		Data:     tokenBytes,
	}

	_, err = scheme.ExtractClaims(&token, string(taEndValBytes))

	assert.EqualError(t, err, `scheme-aws-nitro.Scheme.ExtractVerifiedClaims call to AuthenticateDocument failed:AuthenticateDocument::Verify failed:verification error`)
}

func Test_ValidateEvidenceIntegrity_ok(t *testing.T) {
	privateKey, endCertDer, _, caCertDer, err := generateCertsAndKeys(false, false)
	require.NoError(t, err)

	PCRs, err := generatePCRs()
	require.NoError(t, err)
	userData := generateRandomSlice(32)
	nonce := generateRandomSlice(32)
	tokenBytes, err := nitro_enclave_attestation_document.GenerateDocument(PCRs, userData, nonce, endCertDer, [][]byte{caCertDer}, privateKey)
	require.NoError(t, err)

	taEndValBytes, err := genTaEndorsements(caCertDer)
	require.NoError(t, err)

	scheme := &Scheme{}

	token := proto.AttestationToken{
		TenantId: "1",
		Format:   proto.AttestationFormat_AWS_NITRO,
		Data:     tokenBytes,
	}

	err = scheme.ValidateEvidenceIntegrity(&token, string(taEndValBytes), nil)

	assert.NoError(t, err)
}

func Test_AppraiseEvidence_ok(t *testing.T) {
	extractedBytes, err := os.ReadFile("test/extracted.json")
	require.NoError(t, err)

	var ec proto.EvidenceContext
	err = json.Unmarshal(extractedBytes, &ec)
	require.NoError(t, err)

	endorsementsBytes, err := os.ReadFile("test/endorsements.json")
	require.NoError(t, err)

	scheme := &Scheme{}

	attestation, err := scheme.AppraiseEvidence(&ec, []string{string(endorsementsBytes)})
	require.NoError(t, err)

	assert.Equal(t, proto.TrustTier_AFFIRMING, attestation.Result.Status)
}
