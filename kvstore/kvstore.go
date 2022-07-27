// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"errors"
	"fmt"

	"github.com/veraison/services/config"
)

func New(cfg config.Store) (IKVStore, error) {
	if cfg == nil {
		return nil, errors.New("nil configuration")
	}

	backend, err := config.GetString(cfg, DirectiveBackend, nil)
	if err != nil {
		return nil, err
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

	if err := s.Init(cfg); err != nil {
		return nil, err
	}

	return s, nil
}
