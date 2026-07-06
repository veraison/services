// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package arm_cca

import (
	"errors"
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/corim/extensions"
	"github.com/veraison/eat"
	"github.com/veraison/services/scheme/common"
)

const (
	RatsdV2EndorsementProfile = "tag:github.com,2026:veraison/endors/ratsd/v2"
)

func validateEnvironment(env *comid.Environment) error {
	if env.Class == nil {
		return errors.New("class not set in environment")
	}

	if env.Class.Vendor == nil {
		return errors.New("vendor not set in environment")
	}
	if env.Class.Model == nil {
		return errors.New("model not set in environment")
	}

	if env.Group != nil {
		return errors.New("group set in environment")
	}

	return nil
}

func validateCryptoKeys(keys []*comid.CryptoKey) error {

	for _, key := range keys {
		if key.Type() != comid.PKIXBase64CertPathType && key.Type() != comid.PKIXBase64CertType {
			return fmt.Errorf("key must be a cert or a cert path, found: %s", key.Type())
		}
	}
	return nil
}

func validateMeasurements(measurements []comid.Measurement) error {
	if len(measurements) != 1 {
		return fmt.Errorf("only one measurement expected, found: %d", len(measurements))
	}

	return nil
}

func init() {
	profileID, err := eat.NewProfile(RatsdV2EndorsementProfile)
	if err != nil {
		panic(err)
	}

	validator := &common.TriplesValidator{
		TAEnviromentValidator: func(e *comid.Environment) error {
			return validateEnvironment(e)
		},
		RefValEnviromentValidator: func(e *comid.Environment) error {
			return validateEnvironment(e)
		},
		CryptoKeysValidator:   validateCryptoKeys,
		MeasurementsValidator: validateMeasurements,
	}

	extMap := extensions.NewMap().Add(comid.ExtTriples, validator)
	if err := corim.RegisterProfile(profileID, extMap); err != nil {
		panic(err)
	}
}
