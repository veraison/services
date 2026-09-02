// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package coserv

import (
	"time"

	"github.com/veraison/corim/comid"
	"github.com/veraison/services/plugin"
)

// StoreConfig contains the CoSERV related config parameters
// that the endorsement store uses.
type StoreConfig struct {
	Authority *comid.CryptoKey
	MaxExpiry time.Duration
}

func CreateStoreConfig(auth *comid.CryptoKey, exp time.Duration) StoreConfig {
	return StoreConfig{
		auth,
		exp,
	}
}

// private constants used for serializing and deserializing
// StoreConfig
var (
	fallbackAuthorityKey = "coserv-fallback-authority"
	maxExpiryKey         = "coserv-max-expiry"
)

// Deserialize plugin.Parameters into StoreConfig
func (o *StoreConfig) FromParams(params *plugin.Parameters) error {
	var key comid.CryptoKey

	auth, err := params.GetBytes(fallbackAuthorityKey)
	if err != nil {
		return err
	}

	exp, err := params.GetInt64(maxExpiryKey)
	if err != nil {
		return err
	}

	if err := (&key).UnmarshalCBOR(auth); err != nil {
		return err
	}

	o.Authority = &key
	o.MaxExpiry = time.Duration(exp)
	return nil
}

// Serialize StoreConfig to plugin.Parameters
func (o *StoreConfig) ToParams() (*plugin.Parameters, error) {
	auth, err := o.Authority.MarshalCBOR()
	if err != nil {
		return nil, err
	}
	exp := o.MaxExpiry.Nanoseconds()

	params := plugin.NewParameters().
		SetBytes(fallbackAuthorityKey, auth).
		SetInt64(maxExpiryKey, exp)

	return params, nil
}
