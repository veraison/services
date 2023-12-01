// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package cca_realm

import (
	"encoding/json"
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/services/handler"
)

type CorimExtractor struct{}

func (o CorimExtractor) RefValExtractor(
	rv comid.ReferenceValue,
) ([]*handler.Endorsement, error) {
	var classAttrs ClassAttributes

	if err := classAttrs.FromEnvironment(rv.Environment); err != nil {
		return nil, fmt.Errorf("could not extract Realm class attributes: %w", err)
	}

	// For Realm's we only expect one Reference Value
	rvs := make([]*handler.Endorsement, 0, 1)
	var measurements [][]byte
	var algID string
	for _, m := range rv.Measurements {

		d := m.Val.Digests

		if d == nil {
			return nil, fmt.Errorf("measurement value has no digests")
		}
		k := len(*d)
		if k < 1 {
			return nil, fmt.Errorf("expecting atleast one digest")
		}
		algID = (*d)[0].AlgIDToString()
		measurementValue := (*d)[0].HashValue

		measurements = append(measurements, measurementValue)
	}

	attrs, err := makeRefValAttrs(&classAttrs, algID, measurements)
	if err != nil {
		return nil, fmt.Errorf("attributes error: %w", err)
	}

	ev := &handler.Endorsement{
		Scheme:     SchemeName,
		Type:       handler.EndorsementType_REFERENCE_VALUE,
		Attributes: attrs,
	}

	rvs = append(rvs, ev)

	if len(rvs) == 0 {
		return nil, fmt.Errorf("no measurements found")
	}

	return rvs, nil
}

func makeRefValAttrs(cAttr *ClassAttributes, algID string, measurements [][]byte) (json.RawMessage, error) {

	var attrs = map[string]interface{}{
		"cca-realm.vendor":            cAttr.Vendor,
		"cca-realm.model":             cAttr.Model,
		"cca-realm.id":                cAttr.UUID,
		"cca-realm.alg-id":            algID,
		"cca-realm.measurement-array": measurements,
	}
	data, err := json.Marshal(attrs)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal reference value attributes: %w", err)
	}
	return data, nil
}

func (o CorimExtractor) TaExtractor(
	avk comid.AttestVerifKey,
) (*handler.Endorsement, error) {

	return nil, fmt.Errorf("cca realm endorsements does not have a Trust Anchor")
}
