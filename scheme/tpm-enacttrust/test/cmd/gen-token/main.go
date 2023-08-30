// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"os"

	"github.com/google/go-tpm/tpm2"
	"github.com/google/uuid"
)

type TokenDescription struct {
	NodeID          string `json:"node-id"`
	Digest          []byte `json:"digest"`
	PCRs            []int  `json:"pcrs"`
	FirmwareVersion uint64 `json:"firmware"`
	Algorithm       uint16 `json:"algorithm"`
	Type            uint16 `json:"type"`
}

func makeAttestationData(desc *TokenDescription) (*tpm2.AttestationData, error) {
	data := tpm2.AttestationData{
		Magic:           0xff544347,
		FirmwareVersion: desc.FirmwareVersion,
		Type:            tpm2.TagAttestQuote,
		AttestedQuoteInfo: &tpm2.QuoteInfo{
			PCRSelection: tpm2.PCRSelection{Hash: tpm2.AlgSHA256, PCRs: desc.PCRs},
			PCRDigest:    desc.Digest,
		},
	}

	return &data, nil
}

func readTokenDescription(path string) (*TokenDescription, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var desc TokenDescription
	if err = json.Unmarshal(buf, &desc); err != nil {
		return nil, err
	}

	return &desc, nil
}

func readPrivateKey(path string) (*ecdsa.PrivateKey, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(buf)
	if block == nil || block.Type != "EC PRIVATE KEY" {
		return nil, fmt.Errorf("could not decode EC private key from PEM block: %q", block)
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("could not extract EC private key: %w", err)
	}

	return key, err
}

func main() {
	var keyPath, outPath string
	var badNode bool
	var marshaledNodeID []byte
	flag.StringVar(&keyPath, "key", "key.pem", "Path of the ECDSA key used to sign the token data encoded in PEM.")
	flag.StringVar(&outPath, "out", "quote.bin", "Output path of the generated token.")
	flag.BoolVar(&badNode, "bad-node", false,
		"Allow node-id to not be a valid UUID. If this is set, the bytes of the string will be written as-is, rather than attempting to parse UUID out of it. No length check or any other validation will be performed.")
	flag.Parse()
	descPath := flag.Arg(0)

	key, err := readPrivateKey(keyPath)
	if err != nil {
		fmt.Printf("ERROR: could not read key: %v\n", err)
		os.Exit(1)
	}

	desc, err := readTokenDescription(descPath)
	if err != nil {
		fmt.Printf("ERROR: could not read token description: %v\n", err)
		os.Exit(1)
	}

	if badNode {
		marshaledNodeID = []byte(desc.NodeID)
	} else { // node-id must be a string encoding of a UUID
		nodeID, err := uuid.Parse(desc.NodeID)
		if err != nil {
			fmt.Printf("ERROR: could not parse nodeID: %v\n", err)
			os.Exit(1)
		}
		marshaledNodeID, err = nodeID.MarshalBinary()
		if err != nil {
			fmt.Printf("ERROR: could not mashal nodeID: %v\n", err)
			os.Exit(1)
		}
	}

	d, err := makeAttestationData(desc)
	if err != nil {
		fmt.Printf("ERROR: could not generate attestation data: %v\n", err)
		os.Exit(1)
	}

	attest, err := d.Encode()
	if err != nil {
		fmt.Printf("ERROR: could not encode attestation data: %v\n", err)
		os.Exit(1)
	}

	buff := new(bytes.Buffer)
	endianness := binary.BigEndian

	hash := sha256.Sum256(attest)
	r, s, err := ecdsa.Sign(rand.Reader, key, hash[:])
	if err != nil {
		fmt.Printf("ERROR: could not sign attestation data: %v\n", err)
		os.Exit(1)
	}

	sigStruct := tpm2.Signature{
		Alg: tpm2.AlgECDSA,
		ECC: &tpm2.SignatureECC{HashAlg: tpm2.AlgSHA256, R: r, S: s},
	}
	sig, err := sigStruct.Encode()
	if err != nil {
		fmt.Printf("ERROR: could not encode signature: %v\n", err)
		os.Exit(1)
	}

	if err := binary.Write(buff, endianness, marshaledNodeID); err != nil {
		fmt.Printf("ERROR writing NodeID: %v\n", err)
		os.Exit(1)
	}

	attestLen := uint16(len(attest))
	if err := binary.Write(buff, endianness, attestLen); err != nil {
		fmt.Printf("ERROR writing length: %v\n", err)
		os.Exit(1)
	}

	if err := binary.Write(buff, endianness, attest); err != nil {
		fmt.Printf("ERROR writing TPMS_ATTEST structure: %v\n", err)
		os.Exit(1)
	}

	if err := binary.Write(buff, endianness, sig); err != nil {
		fmt.Printf("ERROR writing signature: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outPath, buff.Bytes(), 0600); err != nil {
		fmt.Printf("ERROR could not write %q: %v\n", outPath, err)
		os.Exit(1)
	}
}
