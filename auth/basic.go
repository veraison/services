// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/veraison/services/log"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type basicAuthUser struct {
	Password string   `mapstructure:"password"`
	Roles    []string `mapstructure:"roles"`
}

func newBasicAuthUser(m map[string]interface{}) (*basicAuthUser, error) {
	var newUser basicAuthUser

	passRaw, ok := m["password"]
	if !ok {
		return nil, errors.New("password not set")
	}

	switch t := passRaw.(type) {
	case string:
		newUser.Password = t
	default:
		return nil, fmt.Errorf("invalid password: expected string found %T", t)
	}

	rolesRaw, ok := m["roles"]
	if ok {
		switch t := rolesRaw.(type) {
		case []string:
			newUser.Roles = t
		case string:
			newUser.Roles = make([]string, 1)
			newUser.Roles[0] = t
		default:
			return nil, fmt.Errorf(
				"invalid roles: expected []string or string, found %T",
				t,
			)
		}
	} else {
		newUser.Roles = make([]string, 0)
	}

	return &newUser, nil
}

type BasicAuthorizer struct {
	logger *zap.SugaredLogger
	users  map[string]*basicAuthUser
}

func (o *BasicAuthorizer) Init(v *viper.Viper, logger *zap.SugaredLogger) error {
	if logger == nil {
		return errors.New("nil logger")
	}
	o.logger = logger

	o.users = make(map[string]*basicAuthUser)
	if rawUsers := v.GetStringMap("users"); rawUsers != nil {
		for name, rawUser := range rawUsers {
			switch t := rawUser.(type) {
			case map[string]interface{}:
				newUser, err := newBasicAuthUser(t)
				if err != nil {
					return fmt.Errorf("invalid user %q: %w", name, err)

				}
				o.logger.Debugw("registered user",
					"user", name,
					"password", newUser.Password,
					"roles", newUser.Roles,
				)
				o.users[name] = newUser
			default:
				return fmt.Errorf(
					"invalid user %q: expected map[string]interface{}, got %T",
					name, t,
				)
			}
		}
	}

	return nil
}

func (o *BasicAuthorizer) Close() error {
	return nil
}

func (o *BasicAuthorizer) GetGinHandler(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		o.logger.Debugw("auth basic", "path", c.Request.URL.Path)

		userName, password, hasAuth := c.Request.BasicAuth()
		if !hasAuth {
			c.Writer.Header().Set("WWW-Authenticate", "Basic realm=veraison")
			ReportProblem(c, http.StatusUnauthorized,
				"no Basic Authorizaiton given")
			return
		}

		userInfo, ok := o.users[userName]
		if !ok {
			c.Writer.Header().Set("WWW-Authenticate", "Basic realm=veraison")
			ReportProblem(c, http.StatusUnauthorized,
				"no Basic Authorizaiton given")
			return
		}

		if err := bcrypt.CompareHashAndPassword(
			[]byte(userInfo.Password),
			[]byte(password),
		); err != nil {
			o.logger.Debugf("password check failed: %v", err)
			c.Writer.Header().Set("WWW-Authenticate", "Basic realm=veraison")
			ReportProblem(c, http.StatusUnauthorized,
				"wrong username or password")
			return
		}

		gotRole := false
		if role == NoRole {
			gotRole = true
		} else {
			for _, userRole := range userInfo.Roles {
				if userRole == role {
					gotRole = true
					break
				}
			}
		}

		if gotRole {
			log.Debugw("user authenticated", "user", userName, "role", role)
		} else {
			c.Writer.Header().Set("WWW-Authenticate", "Basic realm=veraison")
			ReportProblem(c, http.StatusUnauthorized,
				"API unauthorized for user")
		}
	}
}
