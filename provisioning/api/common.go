// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/fxamacker/cbor/v2"
	"github.com/gin-gonic/gin"
	"github.com/moogar0880/problems"
	"github.com/veraison/services/log"
)

func ReportProblem(c *gin.Context, status int, details ...string) {
	prob := problems.NewStatusProblem(status)

	if len(details) > 0 {
		prob.Detail = strings.Join(details, ", ")
	}

	log.LogProblem(log.Named("api"), prob)

	c.Header("Content-Type", "application/problem+json")
	c.AbortWithStatusJSON(status, prob)
}

func ReportConciseProblem(c *gin.Context, status int, details ...string) {
	prob := &log.ConciseProblem{
		Title: fmt.Sprintf("%d %v", status, http.StatusText(status)),
	}

	if len(details) > 0 {
		prob.Detail = strings.Join(details, ", ")
	}

	logger := log.Named("api")

	log.LogConciseProblem(logger, status, prob)

	b, err := cbor.Marshal(prob)
	if err != nil {
		log.Error(logger, "failed to marshal problem details to CBOR", "error", err)
		c.AbortWithStatus(status)
	}

	c.Data(status, "application/concise-problem-details+cbor", b)
	c.Abort()
}
