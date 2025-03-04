// Copyright 2023-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package api

import (
	"path"

	"github.com/gin-gonic/gin"
	"github.com/veraison/services/auth"
)

const (
	managementPath = "/management/v1"
)

var publicApiMap = make(map[string]string)

func NewRouter(handler Handler, authorizer auth.IAuthorizer) *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/.well-known/veraison/management", handler.GetManagementWellKnownInfo)

	manageGroup := router.Group(managementPath)
	manageGroup.Use(authorizer.GetGinHandler(auth.ManagerRole))

	manageGroup.POST("policy/:scheme", handler.CreatePolicy)
	publicApiMap["createPolicy"] = path.Join(managementPath, "policy/:scheme")

	manageGroup.POST("policy/:scheme/:uuid/activate", handler.Activate)
	publicApiMap["activatePolicy"] = path.Join(managementPath, "policy/:scheme/:uuid/activate")

	manageGroup.GET("policy/:scheme", handler.GetActivePolicy)
	publicApiMap["getActivePolicy"] = path.Join(managementPath, "policy/:scheme")

	manageGroup.GET("policy/:scheme/:uuid", handler.GetPolicy)
	publicApiMap["getPolicy"] = path.Join(managementPath, "policy/:scheme/:uuid")

	manageGroup.POST("policies/:scheme/deactivate", handler.DeactivateAll)
	publicApiMap["deactivatePolicies"] = path.Join(managementPath,
						"policies/:scheme/deactivate")

	manageGroup.GET("policies/:scheme", handler.GetPolicies)
	publicApiMap["getPolicies"] = path.Join(managementPath, "policies/:scheme")

	return router
}
