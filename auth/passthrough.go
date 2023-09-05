// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type PassthroughAuthorizer struct {
	logger *zap.SugaredLogger
}

func NewPassthroughAuthorizer(logger *zap.SugaredLogger) IAuthorizer {
	return &PassthroughAuthorizer{logger: logger}
}

func (o *PassthroughAuthorizer) Init(v *viper.Viper, logger *zap.SugaredLogger) error {
	if logger == nil {
		return errors.New("nil logger")
	}
	o.logger = logger
	return nil
}

func (o *PassthroughAuthorizer) Close() error {
	return nil
}

func (o *PassthroughAuthorizer) GetGinHandler(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		o.logger.Debugw("passthrough", "path", c.Request.URL.Path)
	}
}
