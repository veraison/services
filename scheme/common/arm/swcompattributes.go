// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package arm

import (
	"fmt"

	"github.com/veraison/corim/comid"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

type SwCompAttributes struct {
	MeasurementType  string
	Version          string
	SignerID         []byte
	AlgID            uint64
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

	o.AlgID = (*d)[0].HashAlgID
	o.MeasurementValue = (*d)[0].HashValue

	return nil
}

func (o SwCompAttributes) GetRefValType() string {
	return "sw-component"
}

func (o *SwCompAttributes) MakeRefAttrs(c ClassAttributes, scheme string) (*structpb.Struct, error) {
	swAttrs := map[string]interface{}{
		scheme + ".impl-id":           c.ImplID,
		scheme + ".signer-id":         o.SignerID,
		scheme + ".measurement-value": o.MeasurementValue,
		scheme + ".measurement-desc":  o.AlgID,
	}

	if c.Vendor != "" {
		swAttrs[scheme+".hw-vendor"] = c.Vendor
	}

	if c.Model != "" {
		swAttrs[scheme+".hw-model"] = c.Model
	}

	if o.MeasurementType != "" {
		swAttrs[scheme+".measurement-type"] = o.MeasurementType
	}

	if o.Version != "" {
		swAttrs[scheme+".version"] = o.Version
	}

	return structpb.NewStruct(swAttrs)
}
