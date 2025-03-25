// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package api

import (
	"path"

	"github.com/gin-gonic/gin"
)

const (
	edApiPath = "/endorsement-distribution/v1"
)

var publicApiMap = make(map[string]string)

func NewRouter(handler Handler) *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/.well-known/veraison/endorsement-distribution", handler.GetEdApiWellKnownInfo)

	coservEndpoint := path.Join(edApiPath, "coserv/:query")
	router.GET(coservEndpoint, handler.CoservRequest)
	publicApiMap["coservRequest"] = coservEndpoint

	return router
}
