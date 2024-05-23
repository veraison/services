// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package arm

import (
	"encoding/json"
	"fmt"

	"github.com/veraison/corim/comid"
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

func (o CCAPlatformConfigID) GetRefValType() string {
	return "platform-config"
}

// For CCAPlatformConfigID object, scheme argument is not strictly required, but is required for other
// usage of the same interface
func (o CCAPlatformConfigID) MakeRefAttrs(c ClassAttributes, scheme string) (json.RawMessage, error) {
	refAttrs := map[string]interface{}{
		scheme + ".impl-id":               c.ImplID,
		scheme + ".platform-config-label": o.Label,
		scheme + ".platform-config-id":    o.Value,
	}

	if c.Vendor != "" {
		refAttrs[scheme+".hw-vendor"] = c.Vendor
	}

	if c.Model != "" {
		refAttrs[scheme+".hw-model"] = c.Model
	}

	msg, err := json.Marshal(refAttrs)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal CCA platform configuration attributes: %w", err)
	}
	return msg, nil
}
