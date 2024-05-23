// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package arm

import (
	"encoding/json"
	"fmt"

	"github.com/veraison/corim/comid"
)

type SwCompAttributes struct {
	MeasurementType  string
	Version          string
	SignerID         []byte
	AlgID            string
	MeasurementValue []byte
}

func (o *SwCompAttributes) FromMeasurement(m comid.Measurement) error {
	if m.Key == nil {
		return fmt.Errorf("measurement key is not present")
	}

	// extract psa-swcomp-id from mkey
	if !m.Key.IsSet() {
		return fmt.Errorf("measurement key is not set")
	}

	id, err := m.Key.GetPSARefValID()
	if err != nil {
		return fmt.Errorf("failed extracting psa-swcomp-id: %w", err)
	}

	o.SignerID = id.SignerID

	if id.Label != nil {
		o.MeasurementType = *id.Label
	}

	if id.Version != nil {
		o.Version = *id.Version
	}

	// extract digest and alg-id from mval
	d := m.Val.Digests

	if d == nil {
		return fmt.Errorf("measurement value has no digests")
	}

	if len(*d) != 1 {
		return fmt.Errorf("expecting exactly one digest")
	}

	o.AlgID = (*d)[0].AlgIDToString()
	o.MeasurementValue = (*d)[0].HashValue

	return nil
}

func (o SwCompAttributes) GetRefValType() string {
	return "sw-component"
}

func (o *SwCompAttributes) MakeRefAttrs(c ClassAttributes, subScheme string) (json.RawMessage, error) {

	swAttrs := map[string]interface{}{
		subScheme + ".impl-id":           c.ImplID,
		subScheme + ".signer-id":         o.SignerID,
		subScheme + ".measurement-value": o.MeasurementValue,
		subScheme + ".measurement-desc":  o.AlgID,
	}

	if c.Vendor != "" {
		swAttrs[subScheme+".hw-vendor"] = c.Vendor
	}

	if c.Model != "" {
		swAttrs[subScheme+".hw-model"] = c.Model
	}

	if o.MeasurementType != "" {
		swAttrs[subScheme+".measurement-type"] = o.MeasurementType
	}

	if o.Version != "" {
		swAttrs[subScheme+".version"] = o.Version
	}
	msg, err := json.Marshal(swAttrs)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal reference attributes: %w", err)
	}
	return msg, nil
}
