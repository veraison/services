// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

func New(v *viper.Viper) (IKVStore, error) {
	if v == nil {
		return nil, errors.New("nil configuration")
	}

	backend := v.GetString(DirectiveBackend)
	if backend == "" {
		return nil, fmt.Errorf("%q not set in config", DirectiveBackend)
	}

	var s IKVStore

	switch backend {
	case "memory":
		s = &Memory{}
	case "sql":
		s = &SQL{}
	default:
		return nil, fmt.Errorf("backend %q is not supported", backend)
	}

	if err := s.Init(v); err != nil {
		return nil, err
	}

	return s, nil
}
