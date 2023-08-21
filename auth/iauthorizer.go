// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// IAuthorizer defines the interface that must be implemented by the veraison
// auth backends.
type IAuthorizer interface {
	// Init initializes the backend based on the configuration inside the
	// provided Viper object and using the provided logger.
	Init(v *viper.Viper, logger *zap.SugaredLogger) error
	// Close terminates the backend. The exact nature of this method is
	// backend-specific.
	Close() error
	// GetGinHandler returns a gin.HandlerFunc that performs authorization
	// based on the specified role. This function can be set as gin
	// middleware by passing it to gin.Engine.Use().
	GetGinHandler(role string) gin.HandlerFunc
}
