// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// ReadRawConfig instantiates a Viper and uses it to read in configuration. If
// path is specified as something other than an empty string, Viper will
// attempt to read it, inferring the format from the file extension. Otherwise,
// it will for a file called "config.yaml" inside current working directory. An
// error will be returned if there are problems reading configuration. Unless
// allowNotFount is set to true, this includes the config file (either the
// explicitly specified one or the implicit config.yaml) not being present.
func ReadRawConfig(path string, allowNotFound bool) (*viper.Viper, error) {
	v := viper.New()

	if path != "" {
		v.SetConfigFile(path)
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		v.AddConfigPath(wd)
		v.SetConfigType("yaml")
		v.SetConfigName("config")
	}

	v.SetEnvPrefix("veraison")
	v.AutomaticEnv()

	err := v.ReadInConfig()
	if errors.As(err, &viper.ConfigFileNotFoundError{}) && allowNotFound {
		err = nil
	}

	return v, err
}

// GetSubs returns a map of name onto the corresponding Sub from the provided
// Viper, ensuring that it is never nil. If the provided Viper does not contain
// the specified Sub (v.Sub(name) returns nil), an error will be returned,
// unless the name is specified as optional by prefixing it with "*", in which
// case a new empty Viper is returned instead.
func GetSubs(v *viper.Viper, names ...string) (map[string]*viper.Viper, error) {
	var missing []string
	subs := make(map[string]*viper.Viper)

	for _, name := range names {
		isOptional := false
		if strings.HasPrefix(name, "*") {
			name = name[1:]
			isOptional = true
		}

		sub := v.Sub(name)
		if sub == nil {
			if isOptional {
				subs[name] = viper.New()
			} else {
				missing = append(missing, name)
			}
		} else {
			subs[name] = sub
		}
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing directives in %s: %s",
			v.ConfigFileUsed(),
			strings.Join(missing, ", "),
		)
	}

	return subs, nil
}
