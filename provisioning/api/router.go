// Copyright 2022-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package api

import (
	"path"

	"github.com/gin-gonic/gin"
	"github.com/veraison/services/auth"
)

var publicApiMap = make(map[string]string)

const (
	provisioningPath           = "/endorsement-provisioning/v1"
	getWellKnownProvisioningInfoPath = "/.well-known/veraison/provisioning"
)

func NewRouter(handler IHandler, authorizer auth.IAuthorizer) *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET(getWellKnownProvisioningInfoPath, handler.GetWellKnownProvisioningInfo)

	provGroup := router.Group(provisioningPath)
	provGroup.Use(authorizer.GetGinHandler(auth.ProvisionerRole))

	provGroup.POST("submit", handler.Submit)
	publicApiMap["provisioningSubmit"] = path.Join(provisioningPath, "submit")


	return router
}
