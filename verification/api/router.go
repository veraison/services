// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package api

import (
	"github.com/gin-gonic/gin"
)

var publicApiMap = make(map[string]string)

const (
	newChallengeResponseSessionUrl  = "/challenge-response/v1/newSession"
	submitEvidenceUrl               = "/challenge-response/v1/session/:id"
	getSessionUrl                   = "/challenge-response/v1/session/:id"
	delSessionUrl                   = "/challenge-response/v1/session/:id"
	getServiceStateUrl              = "/status"
	getWellKnownVerificationInfoUrl = "/.well-known/veraison/verification"
)

func NewRouter(handler IHandler) *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.POST(newChallengeResponseSessionUrl, handler.NewChallengeResponse)
	publicApiMap["newChallengeResponseSession"] = newChallengeResponseSessionUrl

	router.POST(submitEvidenceUrl, handler.SubmitEvidence)

	router.GET(getSessionUrl, handler.GetSession)

	router.DELETE(delSessionUrl, handler.DelSession)

	router.GET(getServiceStateUrl, handler.GetServiceState)

	router.GET(getWellKnownVerificationInfoUrl, handler.GetWellKnownVerificationInfo)

	return router
}
