// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package arm_cca

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/scheme/common"
	"github.com/veraison/swid"
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

	rset := coserv.NewResultSet()

	for i, j := range resultSet {
		var e handler.Endorsement
		err := json.Unmarshal([]byte(j), &e)
		if err != nil {
			return nil, fmt.Errorf("unable to decode result[%d] %q to Endorsement: %w", i, j, err)
		}

		// TODO trust anchors
		if e.Type != "reference value" || e.SubType != "platform.sw-component" {
			log.Warnf("CCA endorsement of type %q and sub-type %q are currently not handled", e.Type, e.SubType)
			continue
		}

		rvt, err := endorsementToReferenceValueTriple(e)
		if err != nil {
			return nil, fmt.Errorf("unable to map result[%d] %q to CoRIM triple: %w", i, j, err)
		}

		rset.AddReferenceValues(*rvt)
	}

	if err := q.AddResults(*rset); err != nil {
		return nil, fmt.Errorf("failure adding the translated result set: %w", err)
	}

	return q.ToCBOR()
}

// TODO move to scheme/common/arm/platform
func endorsementToReferenceValueTriple(e handler.Endorsement) (*comid.ValueTriple, error) {
	var attrs map[string]string

	if err := json.Unmarshal(e.Attributes, &attrs); err != nil {
		return nil, fmt.Errorf("unmarshalling attributes: %w", err)
	}

	// mkey

	signerID, ok := attrs["signer-id"]
	if !ok {
		return nil, errors.New("missing mandatory signer identifier")
	}

	signerIDBytes, err := base64.StdEncoding.DecodeString(signerID)
	if err != nil {
		return nil, fmt.Errorf("decoding signer-id failed: %w", err)
	}

	rvID, err := comid.NewPSARefValID(signerIDBytes)
	if err != nil {
		return nil, fmt.Errorf("instantiating PSA reference value ID: %w", err)
	}

	if label, ok := attrs["measurement-type"]; ok {
		rvID.SetLabel(label)
	}

	if version, ok := attrs["version"]; ok {
		rvID.SetVersion(version)
	}

	// mval

	m, err := comid.NewPSAMeasurement(rvID)
	if err != nil {
		return nil, fmt.Errorf("instantiating PSA measurement: %w", err)
	}

	digest, ok := attrs["measurement-value"]
	if !ok {
		return nil, errors.New("missing mandatory measurement value")
	}

	digestBytes, err := base64.StdEncoding.DecodeString(digest)
	if err != nil {
		return nil, fmt.Errorf("decoding digest failed: %w", err)
	}

	algo, ok := attrs["measurement-desc"]
	if !ok {
		return nil, errors.New("missing mandatory measurement description")
	}

	m.AddDigest(swid.AlgIDFromString(algo), digestBytes)

	measurements := comid.NewMeasurements().Add(m)

	// env

	implID, ok := attrs["impl-id"]
	if !ok {
		return nil, errors.New("missing mandatory implementation identifier")
	}

	implIDBytes, err := base64.StdEncoding.DecodeString(implID)
	if err != nil {
		return nil, fmt.Errorf("decoding implementation identifier failed: %w", err)
	}

	class := comid.NewClassImplID(comid.ImplID(implIDBytes))
	if class == nil {
		return nil, errors.New("class identifier instantiation failed")
	}

	if model, ok := attrs["hw-model"]; ok {
		class.SetModel(model)
	}

	if vendor, ok := attrs["hw-vendor"]; ok {
		class.SetVendor(vendor)
	}

	env := comid.Environment{
		Class: class,
	}

	// rv triple

	return &comid.ValueTriple{
		Environment:  env,
		Measurements: *measurements,
	}, nil
}
