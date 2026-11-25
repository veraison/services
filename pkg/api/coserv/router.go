// Copyright 2025-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package coserv

import (
	"path"

	"github.com/gin-gonic/gin"
)

const (
	edApiPath = "/endorsement-distribution/v1"
)

var publicApiMap = make(map[string]string)

func SetupRouter(router *gin.Engine, handler Handler) *gin.Engine {
	router.GET("/.well-known/coserv-configuration", handler.GetEdApiWellKnownInfo)

	coservEndpoint := path.Join(edApiPath, "coserv/:query")
	publicApiMap["CoSERVRequestResponse"] = path.Join(edApiPath, "coserv/{query}")

	router.GET(coservEndpoint, handler.CoservRequest)

	return router
}
