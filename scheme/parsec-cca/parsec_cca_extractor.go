// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_cca

import (
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/scheme/common/cca/platform"
)

type ParsecCcaExtractor struct {
	Profile string
}

func (o ParsecCcaExtractor) RefValExtractor(
	rvs comid.ValueTriples,
) ([]*handler.Endorsement, error) {
	if o.Profile != "tag:github.com/parallaxsecond,2023-03-03:cca" {
		return nil, fmt.Errorf("invalid profile: %s for scheme PARSEC_CCA", o.Profile)
	}
	subScheme := &platform.CcaSsdExtractor{}
	return subScheme.RefValExtractor(rvs)
}

func (o ParsecCcaExtractor) TaExtractor(avk comid.KeyTriple) (*handler.Endorsement, error) {
	if o.Profile != "tag:github.com/parallaxsecond,2023-03-03:cca" {
		return nil, fmt.Errorf("invalid profile: %s for scheme PARSEC_CCA", o.Profile)
	}
	subScheme := &platform.CcaSsdExtractor{}
	return subScheme.TaExtractor(avk)
}

func (o *ParsecCcaExtractor) SetProfile(profile string) {
	o.Profile = profile
}
