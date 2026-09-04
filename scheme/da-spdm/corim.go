// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package da_spdm

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/corim/extensions"
	"github.com/veraison/services/scheme/common"
)

const ProfileString = "tag:linaro.org,2025:device-spdm#1.0.0"

func validateEnvironment(env *comid.Environment) error {
	if env.Class == nil {
		return errors.New("class not set")
	}

	if env.Class.ClassID == nil || !env.Class.ClassID.IsSet(){
		return errors.New("class ID not set")
	}

	if env.Class.ClassID.Type() != comid.BytesType {
		return fmt.Errorf("class ID: expected bytes, found %s", env.Class.ClassID.Type())
	}

	classIDBytes := env.Class.ClassID.Bytes()
	if !utf8.Valid(classIDBytes) {
		return fmt.Errorf("class ID: must be a valid UTF-8 string")
	}

	if env.Instance != nil {
		return errors.New("instance must not be set")
	}

	if env.Group != nil {
		return errors.New("group must not be set")
	}

	return nil
}

func validateMeasurements(measurements []comid.Measurement) error {
	for i, measurement := range measurements {
		if measurement.Key == nil {
			return fmt.Errorf("measurement[%d]: mkey not set", i)
		}

		if _, err := measurement.Key.GetKeyUint(); err != nil {
			return fmt.Errorf("measurement[%d]: %w", i, err)
		}

		if measurement.Val.Name == nil {
			return fmt.Errorf("measurement[%d]: name not set", i)
		}
		
		if measurement.Val.Digests == nil && measurement.Val.RawValue == nil {
			return fmt.Errorf("measurement[%d]: either digests or raw-value must be set", i)
		}

		if measurement.Val.Digests != nil && measurement.Val.RawValue != nil {
			return fmt.Errorf("measurement[%d]: digests and raw-value can't both be set", i)
		}
	}

	return nil
}

func init() {
	profileID, err := corim.NewProfileFromString(ProfileString)
	if err != nil {
		panic(err)
	}

	validator := &common.TriplesValidator{
		EnviromentValidator: validateEnvironment,
		MeasurementsValidator: validateMeasurements,
	}

	extMap := extensions.NewMap().Add(comid.ExtTriples, validator)
	if err := corim.RegisterProfile(profileID, extMap); err != nil {
		panic(err)
	}
}
