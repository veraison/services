// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package nvidia

import (
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/corim/extensions"
	"github.com/veraison/services/scheme/common"
)

const (
	ProfileString = "2.16.840.1.113741.1.16.2"
)

func validateTrustAnchorEnvironment(env *comid.Environment) error {

	return nil
}

func validateReferenceValueEnvironment(env *comid.Environment) error {

	return nil
}

func validateCryptoKeys(keys []*comid.CryptoKey) error {

	return nil
}

func validateMeasurements(measurements []comid.Measurement) error {

	return nil
}

func init() {
	profileID, err := corim.NewProfileFromString(ProfileString)
	if err != nil {
		panic(err)
	}

	validator := &common.TriplesValidator{
		TAEnviromentValidator:     validateTrustAnchorEnvironment,
		RefValEnviromentValidator: validateReferenceValueEnvironment,
		CryptoKeysValidator:       validateCryptoKeys,
		MeasurementsValidator:     validateMeasurements,
	}
	extMap := extensions.NewMap().Add(comid.ExtTriples, validator)

	if err := corim.RegisterProfile(profileID, extMap); err != nil {
		panic(err)
	}
}
