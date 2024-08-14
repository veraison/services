// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package platform

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/veraison/corim/comid"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/scheme/common/arm/platform"
)

type CcaSsdExtractor struct {
	Scheme string
}

func (o CcaSsdExtractor) RefValExtractor(rvs comid.ValueTriples) ([]*handler.Endorsement, error) {
	refVals := make([]*handler.Endorsement, 0, len(rvs.Values))

	for i, rv := range rvs.Values {
		var classAttrs platform.ClassAttributes
		var refVal *handler.Endorsement
		var err error

		if err := classAttrs.FromEnvironment(rv.Environment); err != nil {
			return nil, fmt.Errorf("could not extract PSA class attributes: %w", err)
		}

		if rv.Measurement.Key == nil {
			return nil, fmt.Errorf("measurement key is not present")
		}

		if !rv.Measurement.Key.IsSet() {
			return nil, fmt.Errorf("measurement key is not set")
		}

		// Check which MKey is present and then decide which extractor to invoke
		switch rv.Measurement.Key.Type() {
		case comid.PSARefValIDType:
			var swCompAttrs platform.SwCompAttributes

			refVal, err = o.extractMeasurement(
				&swCompAttrs,
				rv.Measurement,
				classAttrs,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"unable to extract measurement at index %d, %w",
					i,
					err,
				)
			}
		case comid.CCAPlatformConfigIDType:
			var ccaPlatformConfigID CCAPlatformConfigID
			refVal, err = o.extractMeasurement(
				&ccaPlatformConfigID,
				rv.Measurement,
				classAttrs,
			)
			if err != nil {
				return nil, fmt.Errorf("unable to extract measurement: %w", err)
			}
		default:
			return nil, fmt.Errorf(
				"unknown measurement key: %T",
				reflect.TypeOf(rv.Measurement.Key),
			)
		}
		refVals = append(refVals, refVal)
	}

	if len(refVals) == 0 {
		return nil, fmt.Errorf("no software components found")
	}

	return refVals, nil
}

func (o CcaSsdExtractor) extractMeasurement(
	obj platform.MeasurementExtractor,
	m comid.Measurement,
	class platform.ClassAttributes,
) (*handler.Endorsement, error) {
	if err := obj.FromMeasurement(m); err != nil {
		return nil, err
	}

	refAttrs, err := obj.MakeRefAttrs(class)
	if err != nil {
		return &handler.Endorsement{}, fmt.Errorf("failed to create software component attributes: %w", err)
	}
	refVal := handler.Endorsement{
		Scheme:     o.Scheme,
		Type:       handler.EndorsementType_REFERENCE_VALUE,
		SubType:    obj.GetRefValType(),
		Attributes: refAttrs,
	}
	return &refVal, nil
}

func (o CcaSsdExtractor) TaExtractor(avk comid.KeyTriple) (*handler.Endorsement, error) {
	// extract implementation ID
	var classAttrs platform.ClassAttributes
	if err := classAttrs.FromEnvironment(avk.Environment); err != nil {
		return nil, fmt.Errorf("could not extract PSA class attributes: %w", err)
	}

	// extract instance ID
	var instanceAttrs platform.InstanceAttributes
	if err := instanceAttrs.FromEnvironment(avk.Environment); err != nil {
		return nil, fmt.Errorf("could not extract PSA instance-id: %w", err)
	}

	// extract IAK pub
	if len(avk.VerifKeys) != 1 {
		return nil, errors.New("expecting exactly one IAK public key")
	}

	iakPub := avk.VerifKeys[0]
	if _, ok := iakPub.Value.(*comid.TaggedPKIXBase64Key); !ok {
		return nil, fmt.Errorf("IAK does not appear to be a PEM key (%T)", iakPub.Value)
	}

	taAttrs, err := makeTaAttrs(instanceAttrs, classAttrs, iakPub)
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
	i platform.InstanceAttributes,
	c platform.ClassAttributes,
	key *comid.CryptoKey,
) (json.RawMessage, error) {
	taID := map[string]interface{}{
		"impl-id": c.ImplID,
		"inst-id": []byte(i.InstID),
		"iak-pub": key.String(),
	}

	if c.Vendor != "" {
		taID["hw-vendor"] = c.Vendor
	}

	if c.Model != "" {
		taID["hw-model"] = c.Model
	}

	msg, err := json.Marshal(taID)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal TA attributes: %w", err)
	}
	return msg, nil
}
