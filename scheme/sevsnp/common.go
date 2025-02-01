// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package sevsnp

import (
	"github.com/veraison/corim/comid"
)

// measurementByUintKey looks up comid.Measurement in a CoMID by its MKey.
//
//	If no measurements are found, returns nil and no error. Otherwise,
//	returns the error encountered.
func measurementByUintKey(refVal comid.ValueTriple,
	key uint64) (*comid.Measurement, error) {
	for _, m := range refVal.Measurements.Values {
		if m.Key == nil || !m.Key.IsSet() ||
			m.Key.Type() != comid.UintType {
			continue
		}

		k, err := m.Key.GetKeyUint()
		if err != nil {
			return nil, err
		}

		if k == key {
			return &m, nil
		}
	}

	return nil, nil
}
