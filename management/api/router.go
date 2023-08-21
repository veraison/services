// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"github.com/gin-gonic/gin"
	"github.com/veraison/services/auth"
)

var publicApiMap = map[string]string{
	"createPolicy":       "/management/v1/policy/:scheme",
	"activatePolicy":     "/management/v1/policy/:scheme/:uuid/activate",
	"getActivePolicy":    "/management/v1/policy/:scheme",
	"getPolicy":          "/management/v1/policy/:scheme/:uuid",
	"deactivatePolicies": "/management/v1/policies/:scheme/deactivate",
	"getPolicies":        "/management/v1/policies/:scheme",
}

func NewRouter(handler Handler, authorizer auth.IAuthorizer) *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(authorizer.GetGinHandler(auth.ManagerRole))

	router.POST(publicApiMap["createPolicy"], handler.CreatePolicy)
	router.POST(publicApiMap["activatePolicy"], handler.Activate)
	router.GET(publicApiMap["getActivePolicy"], handler.GetActivePolicy)
	router.GET(publicApiMap["getPolicy"], handler.GetPolicy)

	router.POST(publicApiMap["deactivatePolicies"], handler.DeactivateAll)
	router.GET(publicApiMap["getPolicies"], handler.GetPolicies)

	router.GET("/.well-known/veraison/management", handler.GetManagementWellKnownInfo)

	return router
}
