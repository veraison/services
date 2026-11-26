// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package verification

import (
	"github.com/gin-gonic/gin"
)

var publicApiMap = make(map[string]string)

const (
	newChallengeResponseSessionUrl  = "/challenge-response/v1/newSession"
	submitEvidenceUrl               = "/challenge-response/v1/session/:id"
	getSessionUrl                   = "/challenge-response/v1/session/:id"
	delSessionUrl                   = "/challenge-response/v1/session/:id"
	getWellKnownVerificationInfoUrl = "/.well-known/veraison/verification"
)

func SetupRouter(router *gin.Engine, handler IHandler) *gin.Engine {
	router.POST(newChallengeResponseSessionUrl, handler.NewChallengeResponse)
	publicApiMap["newChallengeResponseSession"] = newChallengeResponseSessionUrl

	router.POST(submitEvidenceUrl, handler.SubmitEvidence)

	router.GET(getSessionUrl, handler.GetSession)

	router.DELETE(delSessionUrl, handler.DelSession)

	router.GET(getWellKnownVerificationInfoUrl, handler.GetWellKnownVerificationInfo)

	return router
}
