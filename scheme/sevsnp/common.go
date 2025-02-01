// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package sevsnp

import (
	"encoding/pem"
	"errors"
	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/proto/sevsnp"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/ratsd/tokens"
	"github.com/veraison/services/proto"
)

var (
	ErrCertificateReadFailure = errors.New("failed to read certificate")
	ErrMissingCertChain       = errors.New("evidence missing certificate chain")
)

// measurementByUintKey looks up comid.Measurement in a CoMID by its MKey.
//
//	If no measurements are found, returns nil and no error. Otherwise,
//	returns the error encountered.
func measurementByUintKey(refVal comid.ValueTriple,
	key uint64) (*comid.Measurement, error) {
	for _, m := range refVal.Measurements.Values {
		if m.Key == nil || !m.Key.IsSet() ||
			m.Key.Type() != comid.UintType {
			continue
		}

		k, err := m.Key.GetKeyUint()
		if err != nil {
			return nil, err
		}

		if k == key {
			return &m, nil
		}
	}

	return nil, nil
}

func parseEvidence(token *proto.AttestationToken) (*tokens.TSMReport, error) {
	var tsm tokens.TSMReport

	err := tsm.FromCBOR(token.Data)
	if err != nil {
		return nil, err
	}

	return &tsm, nil
}

// comidFromJson accepts a CoRIM in JSON format and returns its first CoMID
//
//	Returns error if there are more than a single CoMID, or passes on
//	error from corim routine.
func comidFromJson(buf []byte) (*comid.Comid, error) {
	extractedCorim, err := corim.UnmarshalUnsignedCorimFromJSON(buf)
	if err != nil {
		return nil, err
	}

	if len(extractedCorim.Tags) > 1 {
		return nil, errors.New("too many tags")
	}

	extractedComid, err := corim.UnmarshalComidFromCBOR(
		extractedCorim.Tags[0].Content,
		extractedCorim.Profile,
	)

	if err != nil {
		return nil, err
	}

	return extractedComid, nil
}

func parseCertificateChainFromEvidence(tsm *tokens.TSMReport) (*sevsnp.CertificateChain, error) {
	var certTable abi.CertTable

	if len(tsm.AuxBlob) == 0 {
		return nil, ErrMissingCertChain
	}

	if err := certTable.Unmarshal(tsm.AuxBlob); err != nil {
		return nil, err
	}

	return certTable.Proto(), nil
}

func readCert(cert []byte) ([]byte, error) {
	if len(cert) == 0 {
		return nil, errors.New("empty certificate")
	}

	block, _ := pem.Decode(cert)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, ErrCertificateReadFailure
	}
	return block.Bytes, nil
}
