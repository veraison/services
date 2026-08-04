// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_cca

import (
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/corim/extensions"
	"github.com/veraison/corim/profiles/cca"
)

const ProfileString = "tag:github.com/parallaxsecond,2023-03-03:cca"

func init() {
	profileID, err := corim.NewProfileFromString(ProfileString)
	if err != nil {
		panic(err)
	}

	extMap := extensions.NewMap().Add(comid.ExtTriples, &cca.PlatformTriplesExtensions{})
	if err := corim.RegisterProfile(profileID, extMap); err != nil {
		panic(err)
	}
}
