// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_tpm

import (
	"errors"
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/eat"
	"github.com/veraison/services/proto"
	"github.com/veraison/swid"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

type CorimExtractor struct{}

func (o CorimExtractor) RefValExtractor(
	rv comid.ReferenceValue,
) ([]*proto.Endorsement, error) {
	var id ID

	if err := id.FromEnvironment(rv.Environment); err != nil {
		return nil, fmt.Errorf("could not extract id from ref-val environment: %w", err)
	}

	rvs := make([]*proto.Endorsement, 0, len(rv.Measurements))

	for i, m := range rv.Measurements {
		pcr, err := extractPCR(m)
		if err != nil {
			return nil, fmt.Errorf("could not extract PCR: %w", err)
		}

		digests, err := extractDigests(m)
		if err != nil {
			return nil, fmt.Errorf("measurement[%d]: %w", i, err)
		}

		for j, digest := range digests {
			attrs, err := makeRefValAttrs(id.class, pcr, digest)
			if err != nil {
				return nil, fmt.Errorf("measurement[%d].digest[%d]: %w", i, j, err)
			}

			rv := &proto.Endorsement{
				Scheme:     SchemeName,
				Type:       proto.EndorsementType_REFERENCE_VALUE,
				Attributes: attrs,
			}

			rvs = append(rvs, rv)
		}
	}

	if len(rvs) == 0 {
		return nil, fmt.Errorf("no measurements found")
	}

	return rvs, nil
}

func (o CorimExtractor) TaExtractor(
	avk comid.AttestVerifKey,
) (*proto.Endorsement, error) {
	var id ID

	if err := id.FromEnvironment(avk.Environment); err != nil {
		return nil, fmt.Errorf("could not extract id from AVK environment: %w", err)
	}

	if len(avk.VerifKeys) != 1 {
		return nil, errors.New("expecting exactly one AK public key")
	}

	// Key can't be empty/nil -- the corim decoder is validating this
	akPub := avk.VerifKeys[0].Key

	taAttrs, err := makeTaAttrs(id, akPub)
	if err != nil {
		return nil, fmt.Errorf("failed to create trust anchor raw public key: %w", err)
	}

	ta := &proto.Endorsement{
		Scheme:     SchemeName,
		Type:       proto.EndorsementType_VERIFICATION_KEY,
		Attributes: taAttrs,
	}

	return ta, nil
}

func makeRefValAttrs(class string, pcr uint64, digest swid.HashEntry) (*structpb.Struct, error) {
	return structpb.NewStruct(
		map[string]interface{}{
			"parsec-tpm.class-id": class,
			"parsec-tpm.pcr":      pcr,
			"parsec-tpm.digest":   digest.HashValue,
			"parsec-tpm.alg-id":   digest.HashAlgID,
		},
	)
}

func makeTaAttrs(id ID, key string) (*structpb.Struct, error) {
	if id.instance == nil {
		return nil, errors.New("instance not found in ID")
	}

	attrs := map[string]interface{}{
		"parsec-tpm.class-id":    id.class,
		"parsec-tpm.instance-id": []byte(*id.instance),
		"parsec-tpm.ak-pub":      key,
	}

	return structpb.NewStruct(attrs)
}

func extractPCR(m comid.Measurement) (uint64, error) {
	if m.Key == nil {
		return comid.MaxUint64, fmt.Errorf("measurement key is not present")
	}

	if !m.Key.IsSet() {
		return comid.MaxUint64, fmt.Errorf("measurement key is not set")
	}

	pcr, err := m.Key.GetKeyUint()
	if err != nil {
		return 0, fmt.Errorf("measurement key is not uint: %w", err)
	}

	return pcr, nil
}

func extractDigests(m comid.Measurement) ([]swid.HashEntry, error) {
	if m.Val.Digests == nil {
		return nil, fmt.Errorf("measurement value does not contain digests")
	}

	return *m.Val.Digests, nil
}

type ID struct {
	class    string
	instance *eat.UEID
}

func (o *ID) FromEnvironment(e comid.Environment) error {
	if e.Instance != nil {
		i, err := e.Instance.GetUEID()
		if err != nil {
			return fmt.Errorf("could not extract instance-id (UEID) from instance: %w", err)
		}
		o.instance = &i
	}

	if e.Class == nil {
		return fmt.Errorf("class not found in environment")
	}

	classID := e.Class.ClassID
	if classID == nil {
		return fmt.Errorf("class-id not found in class")
	}

	if classID.Type() != comid.ClassIDTypeUUID {
		return fmt.Errorf("class-id not in UUID format")
	}

	o.class = classID.String()

	return nil
}
