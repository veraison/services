// Copyright 2021-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package tpm_enacttrust

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"

	tpm2 "github.com/google/go-tpm/tpm2"
	uuid "github.com/google/uuid"
)

// Token is the container for the decoded EnactTrust token
type Token struct {
	// NodeId is the identifier of the attesting node.
	NodeId uuid.UUID
	// TPMS_ATTEST decoded from the token
	AttestationData *tpm2.AttestationData
	// Raw token bytes
	Raw []byte
	// TPMT_SIGNATURE decoded from the token
	Signature *tpm2.Signature
}

func (t *Token) Decode(data []byte) error {
	var err error

	// The data structure is
	//	NODE_ID||SIZE||TPMS_ATTEST||TPMT_SIGNATURE
	// With  NODE_ID being 16 bytes, the SIZE 2 bytes, and the size of
	// TPMS_ATTEST is contained in the SIZE.
	// As such the size of the token must be at least 18 bytes to
	// accommodate the first two fixed-sized fields.
	if len(data) < 18 {
		return fmt.Errorf("token too small: found %d bytes, need at least 18", len(data))
	}

	t.NodeId, err = uuid.FromBytes(data[:16])
	if err != nil {
		return fmt.Errorf("could not decode node-id: %w", err)
	}

	size := binary.BigEndian.Uint16(data[16:18])
	if len(data) < int(18+size) {
		return fmt.Errorf("TPMS_ATTEST appears truncated; expected %d, but got %d bytes",
			size, len(data)-18)
	}

	t.Raw = data[18 : 18+size]
	t.AttestationData, err = tpm2.DecodeAttestationData(t.Raw)
	if err != nil {
		return fmt.Errorf("could not decode TPMS_ATTEST: %v", err)
	}

	t.Signature, err = tpm2.DecodeSignature(bytes.NewBuffer(data[18+size:]))
	if err != nil {
		return fmt.Errorf("could not decode TPMT_SIGNATURE: %v", err)
	}

	if err := validateSignatureFormat(t.Signature); err != nil {
		return fmt.Errorf("invalid TPMT_SIGNATURE: %w", err)
	}

	return nil
}

func (t Token) VerifySignature(key *ecdsa.PublicKey) error {
	digest := sha256.Sum256(t.Raw)

	if !ecdsa.Verify(key, digest[:], t.Signature.ECC.R, t.Signature.ECC.S) {
		return fmt.Errorf("failed to verify signature")
	}

	return nil
}

// validateSignatureFormat makes sure the signature has the expected format.
// This does NOT check that the signature is valid for the object being signed.
func validateSignatureFormat(sig *tpm2.Signature) error {
	if sig == nil {
		return errors.New("nil")
	}

	if sig.Alg != tpm2.AlgECDSA {
		return fmt.Errorf("invalid alg: %d (must be %d (ECDSA))", sig.Alg, tpm2.AlgECDSA)
	}

	if sig.ECC == nil {
		return errors.New("missing ECC signature")
	}

	if sig.ECC.HashAlg != tpm2.AlgSHA256 {
		return fmt.Errorf("unsupported hash alg: %d (only %d (SHA256) is supported)",
			sig.ECC.HashAlg, tpm2.AlgSHA256)
	}

	if sig.ECC.R == nil || sig.ECC.S == nil {
		return errors.New("missing ECC signature component(s)")
	}

	return nil
}
