// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package arm

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/veraison/corim/comid"
	"github.com/veraison/services/handler"
)

type Extractor struct {
	Scheme string
}

// MeasurementExtractor is an interface to extract measurements from comid
// to construct Reference Value Endorsements using Reference Value type
type MeasurementExtractor interface {
	FromMeasurement(comid.Measurement) error
	GetRefValType() string
	// MakeRefAttrs is an interface method to populate reference attributes.
	MakeRefAttrs(ClassAttributes, string) (json.RawMessage, error)
}

func (o Extractor) RefValExtractor(rv comid.ReferenceValue) ([]*handler.Endorsement, error) {
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
	refVals := make([]*handler.Endorsement, 0, len(rv.Measurements))
	var refVal *handler.Endorsement
	var err error
	for i, m := range rv.Measurements {
		if m.Key == nil {
			return nil, fmt.Errorf("measurement key is not present")
		}

		if !m.Key.IsSet() {
			return nil, fmt.Errorf("measurement key is not set")
		}

		// Check which MKey is present and then decide which extractor to invoke
		if m.Key.IsPSARefValID() { // nolint:gocritic
			var swCompAttrs SwCompAttributes

			refVal, err = extractMeasurement(&swCompAttrs, m, classAttrs, o.Scheme)
			if err != nil {
				return nil, fmt.Errorf("unable to extract measurement at index %d, %w", i, err)
			}
		} else if m.Key.IsCCAPlatformConfigID() {
			if (o.Scheme != "CCA_SSD_PLATFORM") && (o.Scheme != "PARSEC_CCA") {
				return nil, fmt.Errorf("measurement error at index %d: incorrect profile %s", i, o.Scheme)
			}
			var ccaPlatformConfigID CCAPlatformConfigID
			refVal, err = extractMeasurement(&ccaPlatformConfigID, m, classAttrs, o.Scheme)
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

func extractMeasurement(
	obj MeasurementExtractor,
	m comid.Measurement,
	class ClassAttributes,
	scheme string,
) (*handler.Endorsement, error) {
	if err := obj.FromMeasurement(m); err != nil {
		return nil, err
	}

	refAttrs, err := obj.MakeRefAttrs(class, scheme)
	if err != nil {
		return &handler.Endorsement{}, fmt.Errorf("failed to create software component attributes: %w", err)
	}
	refVal := handler.Endorsement{
		Scheme:     scheme,
		Type:       handler.EndorsementType_REFERENCE_VALUE,
		SubType:    scheme + "." + obj.GetRefValType(),
		Attributes: refAttrs,
	}
	return &refVal, nil
}

func (o Extractor) TaExtractor(avk comid.AttestVerifKey) (*handler.Endorsement, error) {
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

	taAttrs, err := makeTaAttrs(instanceAttrs, classAttrs, iakPub, o.Scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to create trust anchor attributes: %w", err)
	}

	// note we do not need a subType for TA
	ta := &handler.Endorsement{
		Scheme:     o.Scheme,
		Type:       handler.EndorsementType_VERIFICATION_KEY,
		Attributes: taAttrs,
	}

	return ta, nil
}

func makeTaAttrs(
	i InstanceAttributes,
	c ClassAttributes,
	key string,
	scheme string,
) (json.RawMessage, error) {
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

	msg, err := json.Marshal(taID)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal TA attributes: %w", err)
	}
	return msg, nil
}
