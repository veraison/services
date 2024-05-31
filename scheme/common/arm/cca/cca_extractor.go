// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package cca

import (
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/services/handler"
)

type CcaExtractor struct {
	Scheme  string
	Profile string
}

func (o CcaExtractor) RefValExtractor(rv comid.ReferenceValue) ([]*handler.Endorsement, error) {

	switch o.Scheme {
	case "CCA":
		switch o.Profile {
		case "http://arm.com/cca/ssd/1":
			subScheme := &CcaSsdExtractor{Scheme: o.Scheme, SubScheme: "CCA_SSD_PLATFORM"}
			return subScheme.RefValExtractor(rv)
		case "http://arm.com/cca/realm/1":
			subScheme := &CcaRealmExtractor{Scheme: o.Scheme, SubScheme: "CCA_REALM"}
			return subScheme.RefValExtractor(rv)
		default:
			return nil, fmt.Errorf("invalid profile: %s, for Scheme: %s", o.Profile, o.Scheme)
		}
	case "PARSEC_CCA":
		if o.Profile == "tag:github.com/parallaxsecond,2023-03-03:cca" {
			subScheme := &CcaSsdExtractor{Scheme: o.Scheme, SubScheme: "CCA_SSD_PLATFORM"}
			return subScheme.RefValExtractor(rv)

		} else {
			return nil, fmt.Errorf("invalid profile: %s for Scheme: %s", o.Profile, o.Scheme)
		}
	default:
		return nil, fmt.Errorf("invalid Scheme: %s", o.Scheme)
	}
}

func (o CcaExtractor) TaExtractor(avk comid.AttestVerifKey) (*handler.Endorsement, error) {
	switch o.Scheme {
	case "CCA":
		switch o.Profile {
		case "http://arm.com/cca/ssd/1":
			subScheme := &CcaSsdExtractor{Scheme: o.Scheme, SubScheme: "CCA_SSD_PLATFORM"}
			return subScheme.TaExtractor(avk)
		case "http://arm.com/cca/realm/1":
			return nil, fmt.Errorf("incorrect Trust Anchor Extractor invoked for Profile: %s", o.Profile)
		default:
			return nil, fmt.Errorf("invalid profile for Scheme: %s", o.Scheme)
		}
	case "PARSEC_CCA":
		if o.Profile == "tag:github.com/parallaxsecond,2023-03-03:cca" {
			subScheme := &CcaSsdExtractor{Scheme: o.Scheme, SubScheme: "CCA_SSD_PLATFORM"}
			return subScheme.TaExtractor(avk)
		} else {
			return nil, fmt.Errorf("invalid profile for Scheme: %s", o.Scheme)
		}
	default:
		return nil, fmt.Errorf("invalid Scheme: %s", o.Scheme)
	}

}

func (o *CcaExtractor) SetProfile(profile string) {
	o.Profile = profile
}
