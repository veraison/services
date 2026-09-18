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
	"go.uber.org/zap"
)

// ConciseProblem is a representation of the problem details structure defined
// in RFC 9290. The response encoding format is CBOR.
type ConciseProblem struct {
	Title  string `cbor:"-1,keyasint"`
	Detail string `cbor:"-2,keyasint,omitempty"`
}

// LogConciseProblem logs a ConciseProblem reported by the Endorsement Lifecycle
// Management (ELM) API (for now). It's placed here instead of log/log.go to avoid
// import cycle between provisioning and log packages.
// 500 problems are logged as errors and the rest as warnings.
// TODO: move this to log/log.go as part of the code restructuring effort
// See: https://github.com/veraison/services/issues/192
func LogConciseProblem(logger *zap.SugaredLogger, status int, prob *ConciseProblem) {
	var logFunc func(msg string, args ...interface{})

	if status >= 500 {
		logFunc = logger.Errorw
	} else {
		logFunc = logger.Warnw
	}

	logFunc("problem encountered", "title", prob.Title, "detail", prob.Detail)
}

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
	prob := &ConciseProblem{
		Title: fmt.Sprintf("%d %v", status, http.StatusText(status)),
	}

	if len(details) > 0 {
		prob.Detail = strings.Join(details, ", ")
	}

	logger := log.Named("api")

	LogConciseProblem(logger, status, prob)

	b, err := cbor.Marshal(prob)
	if err != nil {
		log.Error(logger, "failed to marshal problem details to CBOR", "error", err)
		c.AbortWithStatus(status)
		return
	}

	c.Data(status, "application/concise-problem-details+cbor", b)
	c.Abort()
}
