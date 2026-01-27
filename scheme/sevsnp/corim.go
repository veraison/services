// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package sevsnp

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
	ProfileString    = "tag:amd.com,2024:snp-corim-profile"
	ArkProfileString = "https://amd.com/ark"
)

func validateTrustAnchorEnvironment(env *comid.Environment) error {
	if env.Class == nil {
		return errors.New("missing class")
	}

	if env.Class.Vendor == nil {
		return errors.New("missing vendor")
	}

	if env.Class.Model == nil {
		return errors.New("missing model")
	}

	return nil
}

func validateReferenceValueEnvironment(env *comid.Environment) error {
	if env.Class == nil {
		return errors.New("missing class")
	}

	if env.Class.ClassID == nil {
		return errors.New("missing class ID")
	}

	if env.Class.ClassID.Type() != comid.OIDType {
		return fmt.Errorf("class ID: expected OID, got %s", env.Class.ClassID.Type())
	}

	if env.Instance == nil {
		return errors.New("missing instance")
	}

	if env.Instance.Type() != comid.BytesType {
		return fmt.Errorf("instance: expected bytes, got %s", env.Instance.Type())
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
	for i, mea := range measurements {
		if mea.Key == nil {
			return fmt.Errorf("measurement %d: mkey not set", i)
		}

		if mea.Key.Type() != comid.UintType {
			return fmt.Errorf("measurement %d: mkey type: expected uint, got %s", i, mea.Key.Type())
		}
	}

	return nil
}

func init() {
	profileID, err := eat.NewProfile(ProfileString)
	if err != nil {
		panic(err)
	}

	arkProfileID, err := eat.NewProfile(ArkProfileString)
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

	if err := corim.RegisterProfile(arkProfileID, extMap); err != nil {
		panic(err)
	}
}
