// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package arm

import (
	"errors"
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/scheme/common"
)

type CcaExtractor struct {
	Scheme  string
	Profile string
}

func (o CcaExtractor) RefValExtractor(rv comid.ReferenceValue) ([]*handler.Endorsement, error) {
	var subScheme common.IExtractor
	switch o.Scheme {
	case "CCA_SSD":
		switch o.Profile {
		case "http://arm.com/cca/ssd/1":
			subScheme = &CcaSsdExtractor{Scheme: o.Scheme, SubScheme: "CCA_SSD_PLATFORM"}
			subScheme.SetProfile(o.Profile)
		case "http://arm.com/cca/realm/1":
			subScheme = &CcaRealmExtractor{Scheme: o.Scheme, SubScheme: "CCA_REALM"}
			subScheme.SetProfile(o.Profile)
		default:
			return nil, fmt.Errorf("invalid profile: %s, for Scheme: %s", o.Profile, o.Scheme)
		}
	case "PARSEC_CCA":
		if o.Profile == "tag:github.com/parallaxsecond,2023-03-03:cca" {
			subScheme = &CcaSsdExtractor{}
			subScheme.SetProfile(o.Profile)
		} else {
			return nil, fmt.Errorf("invalid profile: %s for Scheme: %s", o.Profile, o.Scheme)
		}
	default:
		return nil, fmt.Errorf("invalid Scheme: %s", o.Scheme)
	}

	return extractRefVal(subScheme, rv)
}

func extractRefVal(subScheme common.IExtractor, rv comid.ReferenceValue) ([]*handler.Endorsement, error) {
	if subScheme == nil {
		return nil, errors.New("nil extractor in extractRefVal")
	}
	return subScheme.RefValExtractor(rv)
}

func (o CcaExtractor) TaExtractor(avk comid.AttestVerifKey) (*handler.Endorsement, error) {
	var subScheme common.IExtractor
	switch o.Scheme {
	case "CCA_SSD":
		switch o.Profile {
		case "http://arm.com/cca/ssd/1":
			subScheme = &CcaSsdExtractor{Scheme: o.Scheme, SubScheme: "CCA_SSD_PLATFORM"}
		case "http://arm.com/cca/realm/1":
			return nil, fmt.Errorf("incorrect Trust Anchor Extractor invoked for Profile: %s", o.Profile)
		default:
			return nil, fmt.Errorf("invalid profile for Scheme: %s", o.Scheme)
		}
	case "PARSEC_CCA":
		if o.Profile == "tag:github.com/parallaxsecond,2023-03-03:cca" {
			subScheme = &CcaSsdExtractor{}
		} else {
			return nil, fmt.Errorf("invalid profile for Scheme: %s", o.Scheme)
		}
	default:
		return nil, fmt.Errorf("invalid Scheme: %s", o.Scheme)
	}

	return extractTA(subScheme, avk)
}

func extractTA(subScheme common.IExtractor, avk comid.AttestVerifKey) (*handler.Endorsement, error) {
	if subScheme == nil {
		return nil, errors.New("nil extractor in extractRefVal")
	}

	return subScheme.TaExtractor(avk)
}

func (o *CcaExtractor) SetProfile(profile string) {
	o.Profile = profile
}
