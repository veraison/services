// Copyright 2023-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package cca

import (
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/scheme/common/cca/platform"
	"github.com/veraison/services/scheme/common/cca/realm"
)

type CorimExtractor struct {
	Profile string
}

func (o CorimExtractor) RefValExtractor(rvs comid.ValueTriples) ([]*handler.Endorsement, error) {
	switch o.Profile {
	case "http://arm.com/cca/ssd/1":
		subScheme := &platform.CcaSsdExtractor{Scheme: SchemeName}
		return subScheme.RefValExtractor(rvs)
	case "http://arm.com/cca/realm/1":
		subScheme := &realm.RealmExtractor{Scheme: SchemeName}
		return subScheme.RefValExtractor(rvs)
	default:
		return nil, fmt.Errorf("invalid profile %s for scheme %s", o.Profile, SchemeName)
	}
}

func (o CorimExtractor) TaExtractor(avk comid.KeyTriple) (*handler.Endorsement, error) {
	switch o.Profile {
	case "http://arm.com/cca/ssd/1":
		subScheme := &platform.CcaSsdExtractor{Scheme: SchemeName}
		return subScheme.TaExtractor(avk)
	default:
		return nil, fmt.Errorf("invalid profile%s for scheme %s", o.Profile, SchemeName)
	}
}

func (o *CorimExtractor) SetProfile(profile string) {
	o.Profile = profile
}
