// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package tpm_enacttrust

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/services/handler"
)

type Extractor struct {
	Profile string
}

func (o *Extractor) SetProfile(p string) {
	o.Profile = p
}

func (o Extractor) RefValExtractor(rvs comid.ValueTriples) ([]*handler.Endorsement, error) {
	if len(rvs.Values) != 1 {
		return nil, fmt.Errorf("expecting one measurement only")
	}

	rv := rvs.Values[0]
	var instanceAttrs InstanceAttributes

	if err := instanceAttrs.FromEnvironment(rv.Environment); err != nil {
		return nil, fmt.Errorf("could not extract instance attributes: %w", err)
	}


	var (
		swComponents []*handler.Endorsement
		swCompAttrs  SwCompAttributes
		measurement  comid.Measurement = rv.Measurement
	)

	if err := swCompAttrs.FromMeasurement(measurement); err != nil {
		return nil, fmt.Errorf("extracting measurement: %w", err)
	}

	swAttrs, err := makeSwAttrs(instanceAttrs, swCompAttrs)
	if err != nil {
		return nil, fmt.Errorf("failed to create software component attributes: %w", err)
	}

	swComponent := handler.Endorsement{
		Scheme:     SchemeName,
		Type:       handler.EndorsementType_REFERENCE_VALUE,
		SubType:    "enacttrust-tpm.sw-component",
		Attributes: swAttrs,
	}

	swComponents = append(swComponents, &swComponent)

	if len(swComponents) == 0 {
		return nil, fmt.Errorf("no software components found")
	}

	return swComponents, nil
}

func makeSwAttrs(i InstanceAttributes, s SwCompAttributes) (json.RawMessage, error) {
	sw := map[string]interface{}{
		"enacttrust-tpm.node-id": i.NodeID,
		"enacttrust-tpm.digest":  s.Digest,
		"enacttrust-tpm.alg-id":  s.AlgID,
	}
	msg, err := json.Marshal(sw)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (o Extractor) TaExtractor(avk comid.KeyTriple) (*handler.Endorsement, error) {
	var instanceAttrs InstanceAttributes

	if err := instanceAttrs.FromEnvironment(avk.Environment); err != nil {
		return nil, fmt.Errorf("could not extract node id: %w", err)
	}

	// extract AK pub
	if len(avk.VerifKeys) != 1 {
		return nil, errors.New("expecting exactly one AK public key")
	}

	akPub := avk.VerifKeys[0]

	if _, ok := akPub.Value.(*comid.TaggedPKIXBase64Key); !ok {
		return nil, fmt.Errorf("ak-pub does not appear to be a PEM key (%T)", akPub.Value)
	}

	if err := akPub.Valid(); err != nil {
		return nil, fmt.Errorf("could not parse ak-pub: %v", err)
	}

	taAttrs, err := makeTaAttrs(instanceAttrs, akPub)
	if err != nil {
		return nil, fmt.Errorf("failed to create trust anchor raw public key: %w", err)
	}

	ta := &handler.Endorsement{
		Scheme:     SchemeName,
		Type:       handler.EndorsementType_VERIFICATION_KEY,
		Attributes: taAttrs,
	}

	return ta, nil
}

func makeTaAttrs(i InstanceAttributes, key *comid.CryptoKey) (json.RawMessage, error) {
	attrs := map[string]interface{}{
		"enacttrust-tpm.node-id": i.NodeID,
		"enacttrust.ak-pub":      key.String(),
	}

	msg, err := json.Marshal(attrs)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
