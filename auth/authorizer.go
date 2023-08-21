// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/veraison/services/config"
	"go.uber.org/zap"
)

type cfg struct {
	Backend        string                 `mapstructure:"backend,omitempty"`
	BackendConfigs map[string]interface{} `mapstructure:",remain"`
}

func NewAuthorizer(v *viper.Viper, logger *zap.SugaredLogger) (IAuthorizer, error) {
	cfg := cfg{
		Backend: "passthrough",
	}

	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromViper(v); err != nil {
		return nil, err
	}

	var a IAuthorizer

	switch cfg.Backend {
	case "none", "passthrough":
		a = &PassthroughAuthorizer{}
	case "basic":
		a = &BasicAuthorizer{}
	case "keycloak":
		a = &KeycloakAuthorizer{}
	default:
		return nil, fmt.Errorf("backend %q is not supported", cfg.Backend)
	}

	if err := a.Init(v, logger); err != nil {
		return nil, err
	}

	return a, nil
}
