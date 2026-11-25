// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

import "github.com/gin-gonic/gin"

// NewRouter creates a new *gin.Engine initialized for use with services.
func NewRouter() *gin.Engine {
	router := gin.Default()

	if err := router.SetTrustedProxies(nil); err != nil {
		// will not error with nil argument
		panic(err)
	}

	return router
}
