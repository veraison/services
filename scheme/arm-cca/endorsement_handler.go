// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package arm_cca

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/scheme/common"
	"github.com/veraison/services/scheme/common/arm"
)

type EndorsementHandler struct{}

func (o EndorsementHandler) Init(params handler.EndorsementHandlerParams) error {
	return nil // no-op
}

func (o EndorsementHandler) Close() error {
	return nil // no-op
}

func (o EndorsementHandler) GetName() string {
	return "unsigned-corim (CCA platform profile)"
}

func (o EndorsementHandler) GetAttestationScheme() string {
	return SchemeName
}

func (o EndorsementHandler) GetSupportedMediaTypes() []string {
	return EndorsementMediaTypes
}

func (o EndorsementHandler) Decode(data []byte) (*handler.EndorsementHandlerResponse, error) {
	return common.UnsignedCorimDecoder(data, &CorimExtractor{})
}

func (o EndorsementHandler) CoservRepackage(query string, resultSet []string) ([]byte, error) {
	var q coserv.Coserv
	if err := q.FromBase64Url(query); err != nil {
		return nil, err
	}

	// add (dummy, for now) authority
	authority, err := comid.NewCryptoKeyTaggedBytes([]byte("dummyauth"))
	if err != nil {
		return nil, fmt.Errorf("unable to map (dummy) authority: %w", err)
	}

	// add (dummy, for now) expiry
	dummyExpiry := time.Now().Add(time.Hour * 1)

	rset := coserv.NewResultSet()
	rset.SetExpiry(dummyExpiry)

	log.Debugf("result set: %v", resultSet)

	for i, j := range resultSet {
		var e handler.Endorsement
		err := json.Unmarshal([]byte(j), &e)
		if err != nil {
			return nil, fmt.Errorf("unable to decode result[%d] %q to Endorsement: %w", i, j, err)
		}

		switch q.Query.ArtifactType {
		// reference values
		case coserv.ArtifactTypeReferenceValues:
			if e.Type != "reference value" {
				log.Errorf("CCA query-result mismatch: want reference value, got %s", e.Type)
				continue
			} else if e.SubType != "platform.sw-component" {
				log.Warnf("CCA reference values of sub-type %q are not currently handled", e.SubType)
				continue
			}

			rvt, err := arm.EndorsementToReferenceValueTriple(e)
			if err != nil {
				return nil, fmt.Errorf("unable to map result[%d] %q to CoRIM reference value triple: %w", i, j, err)
			}

			rvq := &coserv.RefValQuad{
				Authorities: &[]comid.CryptoKey{*authority},
				RVTriple:    rvt,
			}

			rset.AddReferenceValues(*rvq)

		// trust anchors
		case coserv.ArtifactTypeTrustAnchors:
			if e.Type != "trust anchor" {
				log.Errorf("CCA query-result mismatch: want trust anchor, got %s", e.Type)
				continue
			}

			akt, err := arm.EndorsementToAttestationKeyTriple(e)
			if err != nil {
				return nil, fmt.Errorf("unable to map result[%d] %q to CoRIM attest key triple: %w", i, j, err)
			}

			akq := &coserv.AKQuad{
				Authorities: &[]comid.CryptoKey{*authority},
				AKTriple:    akt,
			}

			rset.AddAttestationKeys(*akq)

		default:
			log.Errorf("CCA CoSERV can only deal with reference values and trust anchors at the moment")
			continue
		}
	}

	if err := q.AddResults(*rset); err != nil {
		return nil, fmt.Errorf("failure adding the translated result set: %w", err)
	}

	return q.ToCBOR()
}
