// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package sevsnp

import (
	"encoding/json"

	"github.com/google/go-sev-guest/abi"
	sevsnpParser "github.com/jraman567/go-gen-ref/cmd/sevsnp"
	"github.com/veraison/corim/corim"
	"github.com/veraison/services/proto"
)

// EvidenceHandler implements the IEvidenceHandler interface for SEVSNP
type EvidenceHandler struct {
}

// GetName returns the name of this evidence handler instance
func (o EvidenceHandler) GetName() string {
	return "sevsnp-evidence-handler"
}

// GetAttestationScheme returns the attestation scheme
func (o EvidenceHandler) GetAttestationScheme() string {
	return SchemeName
}

// GetSupportedMediaTypes returns the supported media types for the SEVSNP scheme
func (o EvidenceHandler) GetSupportedMediaTypes() []string {
	return EvidenceMediaTypes
}

func transformEvidenceToCorim(token *proto.AttestationToken) (*corim.UnsignedCorim, error) {
	tsm, err := parseEvidence(token)
	if err != nil {
		return nil, err
	}

	reportProto, err := abi.ReportToProto(tsm.OutBlob)
	if err != nil {
		return nil, err
	}

	evComid, err := sevsnpParser.ReportToComid(reportProto, 0)
	if err != nil {
		return nil, err
	}

	err = evComid.Valid()
	if err != nil {
		return nil, err
	}

	evCorim := corim.UnsignedCorim{}
	evCorim.SetProfile(EndorsementMediaTypeRV)
	evCorim.AddComid(evComid)

	return &evCorim, nil
}

// ExtractClaims converts evidence in tsm-report format to our
// "internal representation", which is in CoRIM format.
func (o EvidenceHandler) ExtractClaims(
	token *proto.AttestationToken,
	_ []string,
) (map[string]interface{}, error) {
	var claimsSet map[string]interface{}

	evCorim, err := transformEvidenceToCorim(token)
	if err != nil {
		return nil, err
	}

	evJson, err := evCorim.ToJSON()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(evJson, &claimsSet)
	if err != nil {
		return nil, err
	}

	return claimsSet, nil
}
