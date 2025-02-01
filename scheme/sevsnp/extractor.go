// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package sevsnp

import (
	"encoding/json"
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/services/handler"
)

type Extractor struct {
	Profile string
}

// RefValExtractor stores the CoMID values triples in the database as-is.
func (o Extractor) RefValExtractor(rvs comid.ValueTriples) ([]*handler.Endorsement, error) {
	refVals := make([]*handler.Endorsement, 0, len(rvs.Values))

	for _, rv := range rvs.Values {
		rvAttrs, err := json.Marshal(&rv)
		if err != nil {
			return nil, err
		}

		refVal := &handler.Endorsement{
			Scheme:     SchemeName,
			Type:       handler.EndorsementType_REFERENCE_VALUE,
			SubType:    "measurements",
			Attributes: rvAttrs,
		}

		refVals = append(refVals, refVal)
	}

	return refVals, nil
}

// TaExtractor Processes the verification keys supplied in the Endorsement
//
// The trust anchor for SEV-SNP is AMD Root Key (ARK). Stores the key triple in the database as-is.
func (o Extractor) TaExtractor(avk comid.KeyTriple) (*handler.Endorsement, error) {
	if len(avk.VerifKeys) > 1 {
		return nil, fmt.Errorf("expecting at most one key, got %d keys", len(avk.VerifKeys))
	}

	taAttrs, err := json.Marshal(&avk)
	if err != nil {
		return nil, err
	}

	ta := &handler.Endorsement{
		Scheme:     SchemeName,
		Type:       handler.EndorsementType_VERIFICATION_KEY,
		Attributes: taAttrs,
	}

	return ta, nil
}

// SetProfile sets the extractor profile
func (o Extractor) SetProfile(profile string) {
	o.Profile = profile //nolint:staticcheck
}
