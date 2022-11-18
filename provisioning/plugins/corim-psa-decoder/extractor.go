// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/veraison/corim/comid"
	"github.com/veraison/services/proto"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

const (
	ccaProfile = "http://arm.com/cca/ssd/1"
)

type Extractor struct {
	Profile string
}

func (o *Extractor) SetProfile(p string) {
	o.Profile = p
}

// MeasurementExtractor is an interface to extract measurements from comid
// and to make ref attributes from them
type MeasurementExtractor interface {
	FromMeasurement(comid.Measurement) error
	MakeRefAttrs(PSAClassAttributes) (*structpb.Struct, error)
}

func (o Extractor) RefValExtractor(rv comid.ReferenceValue) ([]*proto.Endorsement, error) {
	var psaClassAttrs PSAClassAttributes

	if err := psaClassAttrs.FromEnvironment(rv.Environment); err != nil {
		return nil, fmt.Errorf("could not extract PSA class attributes: %w", err)
	}

	// Each measurement is encoded in a measurement-map of a CoMID
	// reference-triple-record.  Since a measurement-map can encode one or more
	// measurements, a single reference-triple-record can carry as many
	// measurements as needed, provided they belong to the same PSA RoT
	// identified in the subject of the "reference value" triple.  A single
	// reference-triple-record SHALL completely describe the updatable PSA RoT.
	refVals := make([]*proto.Endorsement, 0, len(rv.Measurements))
	var refVal *proto.Endorsement
	var err error
	for i, m := range rv.Measurements {
		if m.Key == nil {
			return nil, fmt.Errorf("measurement key is not present")
		}
		if !m.Key.IsSet() {
			return nil, fmt.Errorf("measurement key is not set")
		}
		// Check which MKey is present and then decide which extractor to invoke
		if m.Key.IsPSARefValID() {
			var psaSwCompAttrs PSASwCompAttributes

			refVal, err = ExtractMeas(&psaSwCompAttrs, m, psaClassAttrs)
			if err != nil {
				return nil, fmt.Errorf("unable to extract measurement at index %d, %w", i, err)
			}
		} else if m.Key.IsCCAPlatformConfigID() {
			if o.Profile != ccaProfile {
				return nil, fmt.Errorf("measurement error at index %d: incorrect profile %s", i, o.Profile)
			}
			var ccaPlatformConfigID CCAPlatformConfigID
			refVal, err = ExtractMeas(&ccaPlatformConfigID, m, psaClassAttrs)
			if err != nil {
				return nil, fmt.Errorf("unable to extract measurement: %w", err)
			}
		} else {
			return nil, fmt.Errorf("unknown measurement key: %T", reflect.TypeOf(m.Key))
		}
		refVals = append(refVals, refVal)
	}

	if len(refVals) == 0 {
		return nil, fmt.Errorf("no software components found")
	}

	return refVals, nil
}

func ExtractMeas(obj MeasurementExtractor, m comid.Measurement, class PSAClassAttributes) (*proto.Endorsement, error) {

	if err := obj.FromMeasurement(m); err != nil {
		return nil, err
	}

	refAttrs, err := obj.MakeRefAttrs(class)
	if err != nil {
		return &proto.Endorsement{}, fmt.Errorf("failed to create software component attributes: %w", err)
	}
	refVal := proto.Endorsement{
		Scheme:     proto.AttestationFormat_PSA_IOT,
		Type:       proto.EndorsementType_REFERENCE_VALUE,
		Attributes: refAttrs,
	}
	return &refVal, nil
}

func (o Extractor) TaExtractor(avk comid.AttestVerifKey) (*proto.Endorsement, error) {
	// extract instance ID
	var psaInstanceAttrs PSAInstanceAttributes

	if err := psaInstanceAttrs.FromEnvironment(avk.Environment); err != nil {
		return nil, fmt.Errorf("could not extract PSA instance-id: %w", err)
	}

	// extract implementation ID
	var psaClassAttrs PSAClassAttributes

	if err := psaClassAttrs.FromEnvironment(avk.Environment); err != nil {
		return nil, fmt.Errorf("could not extract PSA class attributes: %w", err)
	}

	// extract IAK pub
	if len(avk.VerifKeys) != 1 {
		return nil, errors.New("expecting exactly one IAK public key")
	}

	iakPub := avk.VerifKeys[0].Key

	// TODO(tho) check that format of IAK pub is as expected

	taAttrs, err := makeTaAttrs(psaInstanceAttrs, psaClassAttrs, iakPub)
	if err != nil {
		return nil, fmt.Errorf("failed to create trust anchor attributes: %w", err)
	}

	ta := &proto.Endorsement{
		Scheme:     proto.AttestationFormat_PSA_IOT,
		Type:       proto.EndorsementType_VERIFICATION_KEY,
		Attributes: taAttrs,
	}

	return ta, nil
}

func makeTaAttrs(i PSAInstanceAttributes, c PSAClassAttributes, key string) (*structpb.Struct, error) {
	taID := map[string]interface{}{
		"psa.impl-id": c.ImplID,
		"psa.inst-id": []byte(i.InstID),
		"psa.iak-pub": key,
	}

	if c.Vendor != "" {
		taID["psa.hw-vendor"] = c.Vendor
	}

	if c.Model != "" {
		taID["psa.hw-model"] = c.Model
	}

	return structpb.NewStruct(taID)
}
