// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	tpm2 "github.com/google/go-tpm/tpm2"
)

// Token is the container for the decoded EnactTrust token
type Token struct {
	// TPMS_ATTEST decoded from the token
	AttestationData *tpm2.AttestationData
	// Raw token bytes
	Raw []byte
	// TPMT_SIGNATURE decoded from the token
	Signature *tpm2.Signature
}

func (t *Token) Decode(data []byte) error {
	// The first two bytes are the SIZE of the following TPMS_ATTEST
	// structure. The following SIZE bytes are the TPMS_ATTEST structure,
	// the remaining bytes are the signature.
	if len(data) < 3 {
		return fmt.Errorf("could not get data size; token too small (%d)", len(data))
	}

	size := binary.BigEndian.Uint16(data[:2])
	if len(data) < int(2+size) {
		return fmt.Errorf("TPMS_ATTEST appears truncated; expected %d, but got %d bytes",
			size, len(data)-2)
	}

	var err error

	t.Raw = data[2 : 2+size]
	t.AttestationData, err = tpm2.DecodeAttestationData(t.Raw)
	if err != nil {
		return fmt.Errorf("could not decode TPMS_ATTEST: %v", err)
	}

	t.Signature, err = tpm2.DecodeSignature(bytes.NewBuffer(data[2+size:]))
	if err != nil {
		return fmt.Errorf("could not decode TPMT_SIGNATURE: %v", err)
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
