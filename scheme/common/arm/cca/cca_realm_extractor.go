// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package cca

import (
	"encoding/json"
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/services/handler"
)

type CcaRealmExtractor struct {
	Scheme    string
	SubScheme string
}

func (o CcaRealmExtractor) RefValExtractor(rv comid.ReferenceValue) ([]*handler.Endorsement, error) {
	var classAttrs RealmClassAttributes
	var instAttrs RealmInstanceAttributes

	if err := classAttrs.FromEnvironment(rv.Environment); err != nil {
		return nil, fmt.Errorf("could not extract Realm class attributes: %w", err)
	}

	if err := instAttrs.FromEnvironment(rv.Environment); err != nil {
		return nil, fmt.Errorf("could not extract Realm instance attributes: %w", err)
	}

	// Measurements are encoded in a measurement-map of a CoMID
	// reference-triple-record.  Since a measurement-map can encode one or more
	// measurements, a single reference-triple-record can carry as many
	// measurements as needed. Hence for a Realm Instance, all the measurements
	// which comprise of both the "rim" & "rem" measurements are carried in an
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
			SubScheme:  o.SubScheme,
			Type:       handler.EndorsementType_REFERENCE_VALUE,
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
		"CCA_REALM.realm-initial-measurement": *rAttr.Rim,
		"CCA_REALM.hash-alg-id":               rAttr.HashAlgID,
		"CCA_REALM.rim":                       *rAttr.Rim,
	}
	if rAttr.Rpv != nil {
		attrs["CCA_REALM.realm-personalization-value"] = *rAttr.Rpv
	}

	if cAttr.Vendor != nil {
		attrs["CCA_REALM.vendor"] = *cAttr.Vendor
	}
	if cAttr.UUID != nil {
		attrs["CCA_REALM.class-id"] = *cAttr.UUID
	}
	if rAttr.Rem[0] != nil {
		attrs["CCA_REALM.rem0"] = *rAttr.Rem[0]
	}
	if rAttr.Rem[1] != nil {
		attrs["CCA_REALM.rem1"] = *rAttr.Rem[1]
	}
	if rAttr.Rem[2] != nil {
		attrs["CCA_REALM.rem2"] = *rAttr.Rem[2]
	}
	if rAttr.Rem[3] != nil {
		attrs["CCA_REALM.rem3"] = *rAttr.Rem[3]
	}

	data, err := json.Marshal(attrs)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal reference value attributes: %w", err)
	}
	return data, nil
}
