// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"errors"
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/services/proto"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

const schemeName = "TPM_ENACTTRUST"

type Extractor struct {
	Profile string
}

func (o *Extractor) SetProfile(p string) {
	o.Profile = p
}

func (o Extractor) RefValExtractor(rv comid.ReferenceValue) ([]*proto.Endorsement, error) {
	var instanceAttrs InstanceAttributes

	if err := instanceAttrs.FromEnvironment(rv.Environment); err != nil {
		return nil, fmt.Errorf("could not extract instance attributes: %w", err)
	}

	if len(rv.Measurements) != 1 {
		return nil, fmt.Errorf("expecting one measurement only")
	}

	var (
		swComponents []*proto.Endorsement
		swCompAttrs  SwCompAttributes
		measurement  comid.Measurement = rv.Measurements[0]
	)

	if err := swCompAttrs.FromMeasurement(measurement); err != nil {
		return nil, fmt.Errorf("extracting measurement: %w", err)
	}

	swAttrs, err := makeSwAttrs(instanceAttrs, swCompAttrs)
	if err != nil {
		return nil, fmt.Errorf("failed to create software component attributes: %w", err)
	}

	swComponent := proto.Endorsement{
		Scheme:     schemeName,
		Type:       proto.EndorsementType_REFERENCE_VALUE,
		SubType:    "enacttrust-tpm.sw-component",
		Attributes: swAttrs,
	}

	swComponents = append(swComponents, &swComponent)

	if len(swComponents) == 0 {
		return nil, fmt.Errorf("no software components found")
	}

	return swComponents, nil
}

func makeSwAttrs(i InstanceAttributes, s SwCompAttributes) (*structpb.Struct, error) {
	return structpb.NewStruct(
		map[string]interface{}{
			"enacttrust-tpm.node-id": i.NodeID,
			"enacttrust-tpm.digest":  s.Digest,
			"enacttrust-tpm.alg-id":  s.AlgID,
		},
	)
}

func (o Extractor) TaExtractor(avk comid.AttestVerifKey) (*proto.Endorsement, error) {
	var instanceAttrs InstanceAttributes

	if err := instanceAttrs.FromEnvironment(avk.Environment); err != nil {
		return nil, fmt.Errorf("could not extract node id: %w", err)
	}

	// extract AK pub
	if len(avk.VerifKeys) != 1 {
		return nil, errors.New("expecting exactly one AK public key")
	}

	akPub := avk.VerifKeys[0].Key

	taAttrs, err := makeTaAttrs(instanceAttrs, akPub)
	if err != nil {
		return nil, fmt.Errorf("failed to create trust anchor raw public key: %w", err)
	}

	ta := &proto.Endorsement{
		Scheme:     schemeName,
		Type:       proto.EndorsementType_VERIFICATION_KEY,
		Attributes: taAttrs,
	}

	return ta, nil
}

func makeTaAttrs(i InstanceAttributes, key string) (*structpb.Struct, error) {
	attrs := map[string]interface{}{
		"enacttrust-tpm.node-id": i.NodeID,
		"enacttrust.ak-pub":      key,
	}

	return structpb.NewStruct(attrs)
}
