// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package tpm_enacttrust

import (
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/corim/extensions"
	"github.com/veraison/eat"
	"github.com/veraison/services/scheme/common"
)

const ProfileString = "https://enacttrust.com/veraison/1.0.0"

func validateEnvironment(env *comid.Environment) error {
	if env.Instance == nil {
		return errors.New("instance not set in environment")
	}

	if env.Instance.Type() != comid.UUIDType {
		return fmt.Errorf("instance: expected uuid, found %s", env.Instance.Type())
	}

	if env.Class != nil {
		return errors.New("class set in environment")
	}

	if env.Group != nil {
		return errors.New("group set in environment")
	}

	return nil
}

func extractEndorsedDigest(measurements []comid.Measurement) ([]byte, error) {
	if measLen := len(measurements); measLen != 1 {
		return nil, fmt.Errorf("expected exactly one measurement, found %d", measLen)
	}

	mea := measurements[0]

	if mea.Val.Digests == nil {
		return nil, errors.New("no digests in measurement")
	}

	if digestLen := len(*mea.Val.Digests); digestLen != 1 {
		return nil, fmt.Errorf("expected exactly one digest in measurement, found %d", digestLen)
	}

	return (*mea.Val.Digests)[0].HashValue, nil
}

func extractKey(keys []*comid.CryptoKey) (*ecdsa.PublicKey, error) {
	keysLen := len(keys)
	if keysLen != 1 {
		return nil, fmt.Errorf("expected trust anchor to contain exactly one key; found %d", keysLen)
	}

	akPub := keys[0]
	if err := akPub.Valid(); err != nil {
		return nil, fmt.Errorf("could not parse ak-pub: %v", err)
	}

	key, err := common.DecodePublicKeyPEM([]byte(akPub.String()))
	if err != nil {
		return nil, err
	}

	ret, ok := key.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("could not extract EC public key; got [%T]: %v", key, err)
	}

	return ret, nil
}

func init() {
	profileID, err := eat.NewProfile(ProfileString)
	if err != nil {
		panic(err)
	}

	validator := &common.TriplesValidator{
		EnviromentValidator: validateEnvironment,
		MeasurementsValidator: func(measurements []comid.Measurement) error {
			_, err := extractEndorsedDigest(measurements)
			return err
		},
		CryptoKeysValidator: func(keys []*comid.CryptoKey) error {
			_, err := extractKey(keys)
			return err
		},
	}

	extMap := extensions.NewMap().Add(comid.ExtTriples, validator)
	if err := corim.RegisterProfile(profileID, extMap); err != nil {
		panic(err)
	}
}
