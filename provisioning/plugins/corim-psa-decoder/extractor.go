// Copyright 2022-2023 Contributors to the Veraison project.
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
	psaProfile    = "http://arm.com/psa/iot/1"
	ccaProfile    = "http://arm.com/cca/ssd/1"
	psaSchemeName = "PSA_IOT"
	ccaSchemeName = "CCA_SSD_PLATFORM"
)

type Extractor struct {
	Profile string
}

func (o *Extractor) SetProfile(p string) {
	o.Profile = p
}

// MeasurementExtractor is an interface to extract measurements from comid
// to construct Reference Value Endorsements using Reference Value type
type MeasurementExtractor interface {
	FromMeasurement(comid.Measurement) error
	GetRefValType() string
	MakeRefAttrs(ClassAttributes, string) (*structpb.Struct, error)
}

func (o Extractor) RefValExtractor(rv comid.ReferenceValue) ([]*proto.Endorsement, error) {
	var classAttrs ClassAttributes

	if err := classAttrs.FromEnvironment(rv.Environment); err != nil {
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
			// Check correct profile and then proceed
			switch o.Profile {
			case psaProfile, ccaProfile:
				break
			default:
				return nil, fmt.Errorf("measurement error at index %d: incorrect profile %s", i, o.Profile)
			}

			var swCompAttrs SwCompAttributes

			refVal, err = extractMeasurement(&swCompAttrs, m, classAttrs, o.Profile)
			if err != nil {
				return nil, fmt.Errorf("unable to extract measurement at index %d, %w", i, err)
			}
		} else if m.Key.IsCCAPlatformConfigID() {
			if o.Profile != ccaProfile {
				return nil, fmt.Errorf("measurement error at index %d: incorrect profile %s", i, o.Profile)
			}
			var ccaPlatformConfigID CCAPlatformConfigID
			refVal, err = extractMeasurement(&ccaPlatformConfigID, m, classAttrs, o.Profile)
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

func extractMeasurement(obj MeasurementExtractor, m comid.Measurement, class ClassAttributes, profile string) (*proto.Endorsement, error) {
	if err := obj.FromMeasurement(m); err != nil {
		return nil, err
	}
	schemeName, scheme, err := profileToSchemeParams(profile)
	if err != nil {
		return nil, err
	}

	refAttrs, err := obj.MakeRefAttrs(class, scheme)
	if err != nil {
		return &proto.Endorsement{}, fmt.Errorf("failed to create software component attributes: %w", err)
	}
	refVal := proto.Endorsement{
		Scheme:     schemeName,
		Type:       proto.EndorsementType_REFERENCE_VALUE,
		SubType:    scheme + "." + obj.GetRefValType(),
		Attributes: refAttrs,
	}
	return &refVal, nil
}

func (o Extractor) TaExtractor(avk comid.AttestVerifKey) (*proto.Endorsement, error) {
	// extract instance ID
	var instanceAttrs InstanceAttributes

	if err := instanceAttrs.FromEnvironment(avk.Environment); err != nil {
		return nil, fmt.Errorf("could not extract PSA instance-id: %w", err)
	}

	// extract implementation ID
	var classAttrs ClassAttributes

	if err := classAttrs.FromEnvironment(avk.Environment); err != nil {
		return nil, fmt.Errorf("could not extract PSA class attributes: %w", err)
	}

	// extract IAK pub
	if len(avk.VerifKeys) != 1 {
		return nil, errors.New("expecting exactly one IAK public key")
	}

	iakPub := avk.VerifKeys[0].Key
	// TODO(tho) check that format of IAK pub is as expected

	schemeName, scheme, err := profileToSchemeParams(o.Profile)
	if err != nil {
		return nil, err
	}

	taAttrs, err := makeTaAttrs(instanceAttrs, classAttrs, iakPub, scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to create trust anchor attributes: %w", err)
	}

	// note we do not need a subType for TA
	ta := &proto.Endorsement{
		Scheme:     schemeName,
		Type:       proto.EndorsementType_VERIFICATION_KEY,
		Attributes: taAttrs,
	}

	return ta, nil
}

func makeTaAttrs(i InstanceAttributes, c ClassAttributes, key string, scheme string) (*structpb.Struct, error) {
	taID := map[string]interface{}{
		scheme + ".impl-id": c.ImplID,
		scheme + ".inst-id": []byte(i.InstID),
		scheme + ".iak-pub": key,
	}

	if c.Vendor != "" {
		taID[scheme+".hw-vendor"] = c.Vendor
	}

	if c.Model != "" {
		taID[scheme+".hw-model"] = c.Model
	}

	return structpb.NewStruct(taID)
}

func profileToSchemeParams(profile string) (string, string, error) {
	// Check correct profile and then proceed
	switch profile {
	case psaProfile:
		return psaSchemeName, "psa", nil
	case ccaProfile:
		return ccaSchemeName, "cca", nil
	default:
		return "", "", fmt.Errorf("could not map profile %s to scheme", profile)
	}
}
