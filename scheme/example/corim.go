// Copyright <TODO> Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package <TODO>

import (
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/corim/extensions"
	"github.com/veraison/eat"
)

const ProfileString = "<TODO>"

type TriplesValidator struct {}

func (o *TriplesValidator) ValidTriples(triples *comid.Triples) error {
	return nil // TODO
}

func init() {
	profileID, err := eat.NewProfile(ProfileString)
	if err != nil {
		panic(err)
	}

	extMap := extensions.NewMap().Add(comid.ExtTriples, &TriplesValidator{})
	if err := corim.RegisterProfile(profileID, extMap); err != nil {
		panic(err)
	}
}
