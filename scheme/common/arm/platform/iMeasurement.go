// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package platform

import (
	"encoding/json"

	"github.com/veraison/corim/comid"
)

// MeasurementExtractor is an interface to extract measurements from comid
// to construct Reference Value Endorsements using Reference Value type
type MeasurementExtractor interface {
	FromMeasurement(comid.Measurement) error
	GetRefValType() string
	// MakeRefAttrs is an interface method to populate reference attributes.
	MakeRefAttrs(ClassAttributes, string) (json.RawMessage, error)
}
