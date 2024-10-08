// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package realm

import (
	"encoding/json"
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/services/handler"
)

type RealmExtractor struct {
	Scheme string
}

func (o RealmExtractor) RefValExtractor(rv comid.ReferenceValue) ([]*handler.Endorsement, error) {
	var classAttrs RealmClassAttributes
	var instAttrs RealmInstanceAttributes

	if err := classAttrs.FromEnvironment(rv.Environment); err != nil {
		return nil, fmt.Errorf("could not extract Realm class attributes: %w", err)
	}

	if err := instAttrs.FromEnvironment(rv.Environment); err != nil {
		return nil, fmt.Errorf("could not extract Realm instance attributes: %w", err)
	}

	// Measurements are encoded in a measurement-map of a CoMID
	// reference-triple-record. For a Realm Instance, all the measurements
	// which comprise both the "RIM" & "REM" measurements are carried in an
	// integrity register
	refVals := make([]*handler.Endorsement, 0, len(rv.Measurements))

	var refVal *handler.Endorsement
	for _, m := range rv.Measurements {
		var rAttr RealmAttributes
		if err := rAttr.FromMeasurement(m); err != nil {
			return nil, fmt.Errorf("unable to extract realm reference attributes from measurement: %w", err)
		}
		refAttrs, err := makeRefValAttrs(&classAttrs, &instAttrs, &rAttr)
		if err != nil {
			return nil, fmt.Errorf("unable to make reference attributes: %w", err)
		}
		refVal = &handler.Endorsement{
			Scheme:     o.Scheme,
			Type:       handler.EndorsementType_REFERENCE_VALUE,
			SubType:    rAttr.GetRefValType(),
			Attributes: refAttrs,
		}
		refVals = append(refVals, refVal)
	}

	if len(refVals) == 0 {
		return nil, fmt.Errorf("no measurements found")
	}

	return refVals, nil
}

func makeRefValAttrs(cAttr *RealmClassAttributes,
	iAttr *RealmInstanceAttributes,
	rAttr *RealmAttributes) (json.RawMessage, error) {

	var attrs = map[string]interface{}{
		"realm-initial-measurement": *rAttr.RIM,
		"hash-alg-id":               rAttr.HashAlgID,
	}
	if rAttr.RPV != nil {
		attrs["realm-personalization-value"] = *rAttr.RPV
	}

	if cAttr.Vendor != nil {
		attrs["vendor"] = *cAttr.Vendor
	}
	if cAttr.UUID != nil {
		attrs["class-id"] = *cAttr.UUID
	}
	if rAttr.REM[0] != nil {
		attrs["rem0"] = *rAttr.REM[0]
	}
	if rAttr.REM[1] != nil {
		attrs["rem1"] = *rAttr.REM[1]
	}
	if rAttr.REM[2] != nil {
		attrs["rem2"] = *rAttr.REM[2]
	}
	if rAttr.REM[3] != nil {
		attrs["rem3"] = *rAttr.REM[3]
	}

	data, err := json.Marshal(attrs)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal reference value attributes: %w", err)
	}
	return data, nil
}
