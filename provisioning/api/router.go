// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package api

import (
	"github.com/gin-gonic/gin"
)

var publicApiMap = make(map[string]string)

const (
	provisioningSubmitUrl           = "/endorsement-provisioning/v1/submit"
	getServiceStateUrl              = "/status"
	getWellKnownProvisioningInfoUrl = "/.well-known/veraison/provisioning"
)

func NewRouter(handler IHandler) *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.POST(provisioningSubmitUrl, handler.Submit)
	publicApiMap["provisioningSubmit"] = provisioningSubmitUrl

	router.GET(getServiceStateUrl, handler.GetServiceState)

	router.GET(getWellKnownProvisioningInfoUrl, handler.GetWellKnownProvisioningInfo)

	return router
}
