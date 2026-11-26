// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package endorsement

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/veraison/corim/comid"
	"github.com/veraison/services/pkg/config"
	"github.com/veraison/services/pkg/fault"
	"github.com/veraison/services/pkg/plugin"
)

func StoreConfigFromViper(v *viper.Viper) (*StoreConfig, error) {
	var ret StoreConfig
	if err := ret.LoadFromViper(v); err != nil {
		return nil, err
	}

	return &ret, nil
}

type StoreConfig struct {
	ActivePlugins    []string
	Signer           ISigner
	PluginParameters map[string]*plugin.Parameters
}

func (o *StoreConfig) LoadFromViper(v *viper.Viper) error {
	var raw rawConfig

	loader := config.NewLoader(&raw)
	if err := loader.LoadFromViper(v); err != nil {
		return err
	}

	if len(raw.ActivePlugins) == 0 {
		return badStoreConfig("no active plugins specified")
	}

	signer, err := NewSigner(raw.Coserv.Signer)
	if err != nil {
		return badStoreConfig("signer: %w", err)
	}

	authority, err := signer.GetAuthority()
	if err != nil {
		return badStoreConfig("authority: %w", err)
	}

	maxExpiry, err := parseDuration(raw.Coserv.MaxExpiry)
	if err != nil {
		return badStoreConfig("max expiry: %w", err)
	}

	coservConfig := CoservConfig{
		Authority: authority,
		MaxExpiry: maxExpiry,
	}

	pluginParameters := make(map[string]*plugin.Parameters, len(raw.PluginParameters))
	for _, name := range raw.ActivePlugins {
		paramsMap, _ := raw.PluginParameters[name]
		params, err := plugin.ParametersFromMap(paramsMap)
		if err != nil {
			return badStoreConfig("plugin %q params: %w", name, err)
		}

		if err := coservConfig.UpdateParams(params); err != nil {
			return badStoreConfig("plugin %q params: %w", name, err)
		}

		pluginParameters[name] = params
	}

	o.ActivePlugins = raw.ActivePlugins
	o.Signer = signer
	o.PluginParameters = pluginParameters

	return nil
}

// CoservConfig contains the CoSERV related config parameters
// that the endorsement store uses.
type CoservConfig struct {
	Authority *comid.CryptoKey
	MaxExpiry time.Duration
}

// private constants used for serializing and deserializing
// StoreConfig
var (
	fallbackAuthorityKey = "coserv-fallback-authority"
	maxExpiryKey         = "coserv-max-expiry"
)

// Deserialize plugin.Parameters into StoreConfig
func (o *CoservConfig) FromParams(params *plugin.Parameters) error {
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

func (o *CoservConfig) UpdateParams(params *plugin.Parameters) error {
	expiry := o.MaxExpiry.Nanoseconds()
	authority, err := o.Authority.MarshalCBOR()
	if err != nil {
		return err
	}

	params.SetBytes(fallbackAuthorityKey, authority).
		SetInt64(maxExpiryKey, expiry)

	return nil
}

type SignerConfig struct {
	Key string `mapstructure:"key"`
	Alg string `mapstructure:"alg"`
}

type rawCoservConfig struct {
	Signer    SignerConfig `mapstructure:"signer"`
	MaxExpiry string       `mapstructure:"max-expiry"`
}

type rawConfig struct {
	Coserv           rawCoservConfig           `mapstructure:"coserv"`
	ActivePlugins    []string                  `mapstructure:"active-plugins"`
	PluginParameters map[string]map[string]any `mapstructure:"plugin-parameters" config:"zerodefault"`
}

func parseDuration(text string) (time.Duration, error) {
	text = strings.ToLower(text)

	re := regexp.MustCompile(`\s*(\d+)\s*(second|sec|s|minute|min|m|hour|h|day|d)s?\s*`)
	matches := re.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return 0, fmt.Errorf("could not match time specifiers in %q", text)
	}

	// Ensure that match indexes cover the entire input; i.e. there is no
	// unrecognized text left after matching.
	indexes := re.FindAllStringIndex(text, -1)
	pos := 0
	for _, idx := range indexes {
		if idx[0] != pos {
			return 0, fmt.Errorf("parsing %q: unrecognized text between indexes %d and %d",
				text, pos, idx[0])
		}

		pos = idx[1]
	}
	if pos != len(text) {
		return 0, fmt.Errorf("parsing %q: unrecognized trailing text", text)
	}

	var ret time.Duration
	for _, match := range matches {
		val, err := strconv.Atoi(match[1])
		if err != nil {
			// cannot get here because we matched \d+
			panic(err)
		}
		duration := time.Duration(val)

		switch match[2] {
		case "s", "sec", "second":
			ret += duration * time.Second
		case "m", "min", "minute":
			ret += duration * time.Minute
		case "h", "hour":
			ret += duration * time.Hour
		case "d", "day":
			ret += duration * 24 * time.Hour
		default:
			// cannot get here because cases exhaust possible matches
			panic(fmt.Errorf("unexpected time period specifier %q", match[2]))
		}
	}

	return ret, nil
}

func badStoreConfig(args ...any) error {
	return fault.BadConfig("endorsement-store", args...)
}
