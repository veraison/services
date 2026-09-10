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

	// CoSERV service is intended as a public API for distributing Endorsements
	// and Reference Values. Therefore, no authentication is added here.
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/.well-known/coserv-configuration", handler.GetEdApiWellKnownInfo)

	coservEndpoint := path.Join(edApiPath, "coserv/:query")
	// use URI template syntax to indicate the variable part in the discovery document
	publicApiMap["CoSERVRequestResponse"] = path.Join(edApiPath, "coserv/{query}")
	router.GET(coservEndpoint, handler.CoservRequest)

	return router
}
