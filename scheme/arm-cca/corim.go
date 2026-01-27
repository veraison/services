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
	LegacyPlatformProfileString = "http://arm.com/cca/ssd/1"
	LegacyRealmProfileString    = "http://arm.com/cca/realm/1"
	PlatformProfileString       = "tag:arm.com,2023:cca_platform#1.0.0"
	RealmProfileString          = "tag:arm.com,2023:realm#1.0.0"
)

func ValidatePlatformEnvironment(env *comid.Environment, isTrustAnchor bool) error {
	if env.Class == nil {
		return errors.New("class not set")
	}

	if env.Class.ClassID == nil {
		return errors.New("class ID not set")
	}

	if env.Class.ClassID.Type() != comid.ImplIDType {
		return fmt.Errorf("class ID: expected psa.impl-id, got %s", env.Class.ClassID.Type())
	}

	if isTrustAnchor {
		if env.Instance == nil {
			return errors.New("instance not set for trust anchor")
		}

		if env.Instance.Type() != comid.UEIDType {
			return fmt.Errorf("instance: expected UEID, got %s", env.Instance.Type())
		}

	} else if env.Instance != nil {
		return errors.New("instance set for reference value")
	}

	return nil
}

func validateRealmEnvironment(env *comid.Environment) error {
	if env.Instance == nil {
		return errors.New("instance not set")
	}

	if env.Instance.Type() != comid.BytesType {
		return fmt.Errorf("instance: expected bytes, got %s", env.Instance.Type())
	}

	return nil
}

func ValidateCryptoKeys(keys []*comid.CryptoKey) error {
	if len(keys) != 1 {
		return fmt.Errorf("expected exactly one key but got %d", len(keys))
	}

	if keys[0].Type() != comid.PKIXBase64KeyType {
		return fmt.Errorf("trust anchor must be a PKIX base64 key, found: %s", keys[0].Type())
	}

	return nil
}

func ValidatePlatformMeasurements(measurements []comid.Measurement) error {
	for i, mea := range measurements {
		if mea.Key == nil {
			return fmt.Errorf("measurement %d key not set", i)
		}

		switch mea.Key.Type() {
		case comid.PSARefValIDType:
			if mea.Val.Digests == nil {
				return fmt.Errorf("measurement %d value: no digests", i)
			}
		case comid.CCAPlatformConfigIDType:
			if mea.Val.RawValue == nil {
				return fmt.Errorf("measurement %d value: no raw value", i)
			}
		default:
			return fmt.Errorf("measurement %d key: unexpected type %s", i, mea.Key.Type())
		}

	}

	return nil
}

func validateRealmMeasurements(measurements []comid.Measurement) error {
	for i, mea := range measurements {
		if mea.Val.RawValue == nil {
			return fmt.Errorf("measurement %d: personalization (raw value) not set", i)
		}

		if mea.Val.IntegrityRegisters == nil {
			return fmt.Errorf("measurement %d integrity registers not set", i)
		}
	}

	return nil
}

func init() {
	platformProfileID, err := eat.NewProfile(PlatformProfileString)
	if err != nil {
		panic(err)
	}

	legacyPlatformProfileID, err := eat.NewProfile(LegacyPlatformProfileString)
	if err != nil {
		panic(err)
	}

	realmProfileID, err := eat.NewProfile(RealmProfileString)
	if err != nil {
		panic(err)
	}

	legacyRealmProfileID, err := eat.NewProfile(LegacyRealmProfileString)
	if err != nil {
		panic(err)
	}

	platformValidator := &common.TriplesValidator{
		TAEnviromentValidator: func(e *comid.Environment) error {
			return ValidatePlatformEnvironment(e, true)
		},
		RefValEnviromentValidator: func(e *comid.Environment) error {
			return ValidatePlatformEnvironment(e, false)
		},
		CryptoKeysValidator:   ValidateCryptoKeys,
		MeasurementsValidator: ValidatePlatformMeasurements,
	}
	platformExtMap := extensions.NewMap().Add(comid.ExtTriples, platformValidator)

	realmValidator := &common.TriplesValidator{
		EnviromentValidator:   validateRealmEnvironment,
		MeasurementsValidator: validateRealmMeasurements,
		DisallowTAs:           true,
	}
	realmExtMap := extensions.NewMap().Add(comid.ExtTriples, realmValidator)

	if err := corim.RegisterProfile(platformProfileID, platformExtMap); err != nil {
		panic(err)
	}

	if err := corim.RegisterProfile(legacyPlatformProfileID, platformExtMap); err != nil {
		panic(err)
	}

	if err := corim.RegisterProfile(realmProfileID, realmExtMap); err != nil {
		panic(err)
	}

	if err := corim.RegisterProfile(legacyRealmProfileID, realmExtMap); err != nil {
		panic(err)
	}
}
