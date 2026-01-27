// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_tpm

import (
	"errors"
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/corim/extensions"
	"github.com/veraison/eat"
	"github.com/veraison/services/scheme/common"
)

const ProfileString = "tag:github.com/parallaxsecond,2023-03-03:tpm"

func validateEnvironment(env *comid.Environment, isTrustAnchor bool) error {
	if env.Class == nil {
		return errors.New("class not set in environment")
	}

	if env.Class.ClassID == nil {
		return errors.New("class ID not set in environment")
	}

	if env.Class.ClassID.Type() != comid.UUIDType {
		return fmt.Errorf("class ID: expected uuid, found %s", env.Class.ClassID.Type())
	}

	if isTrustAnchor {
		if env.Instance == nil {
			return errors.New("instance not set in trust anchor environment")
		}

		if env.Instance.Type() != comid.BytesType {
			return fmt.Errorf("instance: expected bytes, found %s", env.Instance.Type())
		}
	} else if env.Instance != nil {
		return errors.New("instance set in reference value environment")
	}

	if env.Group != nil {
		return errors.New("group set in environment")
	}

	return nil
}

func validateCryptoKeys(keys []*comid.CryptoKey) error {
	if len(keys) != 1 {
		return fmt.Errorf("expected exactly one key but got %d", len(keys))
	}

	if keys[0].Type() != comid.PKIXBase64KeyType {
		return fmt.Errorf("trust anchor must be a PKIX base64 key, found: %s", keys[0].Type())
	}

	return nil
}

func validateMeasurements(measurements []comid.Measurement) error {
	for i, mea := range measurements {
		if mea.Key == nil {
			return fmt.Errorf("measurement %d has no key", i)
		}

		if mea.Key.Type() != comid.UintType {
			return fmt.Errorf("measurement %d key: expected uint, found %s",
				i, mea.Key.Type())

		}

		if mea.Val.Digests == nil {
			return fmt.Errorf("measurement %d does not contain digests", i)
		}
	}

	return nil
}

func init() {
	profileID, err := eat.NewProfile(ProfileString)
	if err != nil {
		panic(err)
	}

	validator := &common.TriplesValidator{
		TAEnviromentValidator: func(e *comid.Environment) error {
			return validateEnvironment(e, true)
		},
		RefValEnviromentValidator: func(e *comid.Environment) error {
			return validateEnvironment(e, false)
		},
		CryptoKeysValidator:   validateCryptoKeys,
		MeasurementsValidator: validateMeasurements,
	}

	extMap := extensions.NewMap().Add(comid.ExtTriples, validator)
	if err := corim.RegisterProfile(profileID, extMap); err != nil {
		panic(err)
	}
}
