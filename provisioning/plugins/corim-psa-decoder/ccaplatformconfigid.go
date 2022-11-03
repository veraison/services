// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"

	"github.com/veraison/corim/comid"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

type CCAPlatformConfigID struct {
	Label string
	Value []byte
}

func (o *CCAPlatformConfigID) FromMeasurement(m comid.Measurement) error {

	id, err := m.Key.GetCCAPlatformConfigID()
	if err != nil {
		return fmt.Errorf("failed extracting mkey for cca-platform-config-id: %w", err)
	}
	o.Label = string(id)

	if m.Val.RawValue == nil {
		return fmt.Errorf("no CCA platform config id set in the measurements")
	}
	r := *m.Val.RawValue

	o.Value, err = r.GetBytes()
	if err != nil {
		return fmt.Errorf("failed to get CCA platform config id: %w", err)
	}
	return nil
}

func (o CCAPlatformConfigID) MakeSwAttrs(c PSAClassAttributes) (*structpb.Struct, error) {
	swAttrs := map[string]interface{}{
		"psa.impl-id":               c.ImplID,
		"cca.platform-config-label": o.Label,
		"cca.platform-config-id":    o.Value,
	}

	if c.Vendor != "" {
		swAttrs["psa.hw-vendor"] = c.Vendor
	}

	if c.Model != "" {
		swAttrs["psa.hw-model"] = c.Model
	}

	return structpb.NewStruct(swAttrs)
}
