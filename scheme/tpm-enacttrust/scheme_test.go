// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package tpm_enacttrust

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"testing"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/veraison/corim/comid"
	"github.com/veraison/services/vts/appraisal"
)

// makeToken assembles an EnactTrust token (NODE_ID||SIZE||TPMS_ATTEST||
// TPMT_SIGNATURE) that binds the given nonce into TPMS_ATTEST.ExtraData and
// signs the TPMS_ATTEST bytes with the given key.
func makeToken(t *testing.T, key *ecdsa.PrivateKey, nonce []byte) []byte {
	t.Helper()

	attest := tpm2.AttestationData{
		Magic:     0xff544347,
		Type:      tpm2.TagAttestQuote,
		ExtraData: nonce,
		AttestedQuoteInfo: &tpm2.QuoteInfo{
			PCRSelection: tpm2.PCRSelection{Hash: tpm2.AlgSHA256, PCRs: []int{1, 2, 3, 4}},
			PCRDigest:    make([]byte, 32),
		},
	}
	attestBytes, err := attest.Encode()
	require.NoError(t, err)

	digest := sha256.Sum256(attestBytes)
	r, s, err := ecdsa.Sign(rand.Reader, key, digest[:])
	require.NoError(t, err)

	sig := tpm2.Signature{
		Alg: tpm2.AlgECDSA,
		ECC: &tpm2.SignatureECC{HashAlg: tpm2.AlgSHA256, R: r, S: s},
	}
	sigBytes, err := sig.Encode()
	require.NoError(t, err)

	nodeID, err := uuid.Parse("7df7714e-aa04-4638-bcbf-434b1dd720f1")
	require.NoError(t, err)
	nodeBytes, err := nodeID.MarshalBinary()
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	require.NoError(t, binary.Write(buf, binary.BigEndian, nodeBytes))
	require.NoError(t, binary.Write(buf, binary.BigEndian, uint16(len(attestBytes))))
	require.NoError(t, binary.Write(buf, binary.BigEndian, attestBytes))
	require.NoError(t, binary.Write(buf, binary.BigEndian, sigBytes))

	return buf.Bytes()
}

// trustAnchor wraps an ECDSA public key as the single verification key of a
// trust anchor, matching what the scheme expects from a CoRIM.
func trustAnchor(t *testing.T, pub *ecdsa.PublicKey) []*comid.KeyTriple {
	t.Helper()

	der, err := x509.MarshalPKIXPublicKey(pub)
	require.NoError(t, err)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})

	key, err := comid.NewPKIXBase64Key(string(pemBytes))
	require.NoError(t, err)

	return []*comid.KeyTriple{{VerifKeys: comid.CryptoKeys{key}}}
}

// TestValidateEvidenceIntegrity_NonceMatch checks that a signed quote whose
// bound nonce equals the session nonce passes the integrity check.
func TestValidateEvidenceIntegrity_NonceMatch(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	nonce := []byte{0x80, 0x0c, 0xc8, 0x5c, 0x41, 0xd2, 0xea, 0x83}
	evidence := &appraisal.Evidence{Data: makeToken(t, key, nonce), Nonce: nonce}

	err = (&Implementation{}).ValidateEvidenceIntegrity(evidence, trustAnchor(t, &key.PublicKey), nil)
	require.NoError(t, err)
}

// TestValidateEvidenceIntegrity_NonceMismatch checks that a correctly-signed but
// stale quote (bound nonce differs from the session nonce) is rejected, so a
// captured quote cannot be replayed. See issue #427.
func TestValidateEvidenceIntegrity_NonceMismatch(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	boundNonce := []byte{0x80, 0x0c, 0xc8, 0x5c, 0x41, 0xd2, 0xea, 0x83}
	sessionNonce := []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}
	evidence := &appraisal.Evidence{Data: makeToken(t, key, boundNonce), Nonce: sessionNonce}

	err = (&Implementation{}).ValidateEvidenceIntegrity(evidence, trustAnchor(t, &key.PublicKey), nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "freshness")
}
