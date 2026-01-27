// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_cca

import (
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/corim/extensions"
	"github.com/veraison/eat"
	arm_cca "github.com/veraison/services/scheme/arm-cca"
	"github.com/veraison/services/scheme/common"
)

const ProfileString = "tag:github.com/parallaxsecond,2023-03-03:cca"

func init() {
	profileID, err := eat.NewProfile(ProfileString)
	if err != nil {
		panic(err)
	}

	validator := &common.TriplesValidator{
		TAEnviromentValidator: func(e *comid.Environment) error {
			return arm_cca.ValidatePlatformEnvironment(e, true)
		},
		RefValEnviromentValidator: func(e *comid.Environment) error {
			return arm_cca.ValidatePlatformEnvironment(e, false)
		},
		CryptoKeysValidator:   arm_cca.ValidateCryptoKeys,
		MeasurementsValidator: arm_cca.ValidatePlatformMeasurements,
	}

	extMap := extensions.NewMap().Add(comid.ExtTriples, validator)
	if err := corim.RegisterProfile(profileID, extMap); err != nil {
		panic(err)
	}
}
