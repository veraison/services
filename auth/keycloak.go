// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package auth

import (
	"errors"
	"flag"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/tbaehler/gin-keycloak/pkg/ginkeycloak"
	"github.com/veraison/services/config"
	"go.uber.org/zap"
	"gopkg.in/square/go-jose.v2/jwt"
)

type keycloakCfg struct {
	Backend string `mapstructure:"backend"`
	Host    string `mapstructure:"host"`
	Port    string `mapstructure:"port"`
	Realm   string `mapstructure:"realm"`
}

type KeycloakAuthorizer struct {
	logger *zap.SugaredLogger
	config ginkeycloak.KeycloakConfig
}

func (o *KeycloakAuthorizer) Init(v *viper.Viper, logger *zap.SugaredLogger) error {
	if logger == nil {
		return errors.New("nil logger")
	}
	o.logger = logger

	// This prevents glog--the logging package used by ginkeycloak--from complaining.
	flag.Parse()

	cfg := keycloakCfg{
		Host:  "localhost",
		Port:  "1111",
		Realm: "veraison",
	}

	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromViper(v); err != nil {
		return err
	}

	o.config = ginkeycloak.KeycloakConfig{
		Url:                fmt.Sprintf("http://%s:%s", cfg.Host, cfg.Port),
		Realm:              cfg.Realm,
		CustomClaimsMapper: mapTenantID,
	}
	return nil
}

func (o *KeycloakAuthorizer) Close() error {
	return nil
}

func (o *KeycloakAuthorizer) GetGinHandler(role string) gin.HandlerFunc {
	return ginkeycloak.Auth(o.getAuthCheck([]string{role}), o.config)
}

func (o *KeycloakAuthorizer) getAuthCheck(
	roles []string,
) ginkeycloak.AccessCheckFunction {
	return func(tc *ginkeycloak.TokenContainer, ctx *gin.Context) bool {
		ctx.Set("token", *tc.KeyCloakToken)
		ctx.Set("uid", tc.KeyCloakToken.PreferredUsername)

		roleOK := ginkeycloak.RealmCheck(roles)(tc, ctx)

		o.logger.Debugw("auth check", "role", roleOK)

		return roleOK
	}
}

func mapTenantID(jsonWebToken *jwt.JSONWebToken, keyCloakToken *ginkeycloak.KeyCloakToken) error {
	var claims map[string]interface{}

	// note: this mapper function will only be called once the JWT had
	// alreadybeen verified by ginkeycloak, so extracting claims without
	// verification here is, in fact, safe.
	if err := jsonWebToken.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return err
	}

	var tenantId string
	rawTenantId, ok := claims["tenant_id"]
	if ok {
		tenantId, ok = rawTenantId.(string)
		if !ok {
			return fmt.Errorf("tenant_id not a string: %v (%T)",
				rawTenantId, rawTenantId)
		}
	} else {
		tenantId = ""
	}

	keyCloakToken.CustomClaims = map[string]string{
		"tenant_id": tenantId,
	}

	return nil
}
