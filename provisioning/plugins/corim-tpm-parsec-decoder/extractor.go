// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"errors"
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/services/proto"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

type Extractor struct{}

func (o Extractor) SwCompExtractor(rv comid.ReferenceValue) ([]*proto.Endorsement, error) {
	return nil, errors.New("not implemented")
}

func (o Extractor) TaExtractor(avk comid.AttestVerifKey) (*proto.Endorsement, error) {
	var instanceAttrs InstanceAttributes

	if err := instanceAttrs.FromEnvironment(avk.Environment); err != nil {
		return nil, fmt.Errorf("could not extract key id: %w", err)
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
		Scheme:     proto.AttestationFormat_TPM_PARSEC,
		Type:       proto.EndorsementType_VERIFICATION_KEY,
		Attributes: taAttrs,
	}

	return ta, nil
}

func makeTaAttrs(i InstanceAttributes, key string) (*structpb.Struct, error) {
	attrs := map[string]interface{}{
		"parsec-tpm.key-id": i.KeyID,
		"parsec-tpm.ak-pub": key,
	}

	return structpb.NewStruct(attrs)
}
