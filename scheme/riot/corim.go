// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package riot

import (
	"errors"
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/corim/extensions"
	"github.com/veraison/eat"
	"github.com/veraison/services/scheme/common"
)


const ProfileString = "tag:veraison-project.com,2026:riot"

func validateEnvironment(env *comid.Environment) error {
	if env.Class.Vendor ==  nil {
		return errors.New("missing vendor")
	}

	if *env.Class.Vendor != "Veraison Project" {
		return errors.New(`vendor must be "Veraison Project"`)
	}

	if env.Class.Model ==  nil {
		return errors.New("missing vendor")
	}

	if *env.Class.Model != "RIOT" {
		return errors.New(`vendor must be "RIOT"`)
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

func init() {
	profileID, err := eat.NewProfile(ProfileString)
	if err != nil {
		panic(err)
	}

	validator := &common.TriplesValidator{
		EnviromentValidator: validateEnvironment,
		CryptoKeysValidator: validateCryptoKeys,
		DisallowRefVals: true,
	}

	extMap := extensions.NewMap().Add(comid.ExtTriples, validator)
	if err := corim.RegisterProfile(profileID, extMap); err != nil {
		panic(err)
	}
}
