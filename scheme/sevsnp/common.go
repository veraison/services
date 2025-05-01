// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package sevsnp

import (
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/google/go-sev-guest/abi"
	"github.com/google/go-sev-guest/proto/sevsnp"
	"github.com/veraison/cmw"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/ratsd/tokens"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/proto"
)

var (
	ErrCertificateReadFailure = errors.New("failed to read certificate")
	ErrMissingCertChain       = errors.New("evidence missing certificate chain")
	ErrMissingCMW             = errors.New("CMW not found in evidence token")
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

func parseAttestationToken(token *proto.AttestationToken) (*tokens.TSMReport, error) {
	var (
		err           error
		tsm           = new(tokens.TSMReport)
		cmwCollection cmw.CMW
	)

	switch token.MediaType {
	case EvidenceMediaTypeTSMCbor:
		err = tsm.FromCBOR(token.Data)
		if err != nil {
			return nil, err
		}
	case EvidenceMediaTypeTSMJson:
		err = tsm.FromJSON(token.Data)
		if err != nil {
			return nil, err
		}
	case EvidenceMediaTypeRATSd:
		eat := make(map[string]interface{})

		err = json.Unmarshal(token.Data, &eat)
		if err != nil {
			return nil, err
		}

		cmwBase64, ok := eat["cmw"].(string)
		if !ok {
			return nil, handler.BadEvidence(ErrMissingCMW)
		}

		cmwJson, err := base64.StdEncoding.DecodeString(cmwBase64)
		if err != nil {
			return nil, err
		}
		
		err = cmwCollection.UnmarshalJSON(cmwJson)
		if err != nil {
			return nil, err
		}

		cmwMonad, err := cmwCollection.GetCollectionItem("tsm-report")
		if err != nil {
			return nil, err
		}

		cmwType, err := cmwMonad.GetMonadType()
		if err != nil {
			return nil, err
		}
		if cmwType != "application/vnd.veraison.configfs-tsm+json" {
			return nil, fmt.Errorf("unexpected CMW type: %s", cmwType)
		}
		cmwValue, err := cmwMonad.GetMonadValue()
		if err != nil {
			return nil, err
		}

		err = tsm.FromJSON(cmwValue)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unexpected media type: %s", token.MediaType)
	}

	return tsm, nil
}
