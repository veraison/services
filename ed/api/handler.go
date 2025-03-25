// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/moogar0880/problems"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/services/capability"
	"github.com/veraison/services/config"
	"github.com/veraison/services/log"
	"go.uber.org/zap"
)

var (
	tenantID       = "0"
	EdApiMediaType = "application/vnd.veraison.edapi+json"
	errTodo        = errors.New("TODO")
)

type Handler struct {
	Logger *zap.SugaredLogger
}

func NewHandler(logger *zap.SugaredLogger) Handler {
	return Handler{
		Logger: logger,
	}
}

func (o Handler) GetEdApiWellKnownInfo(c *gin.Context) {
	offered := c.NegotiateFormat(capability.WellKnownMediaType)
	if offered != capability.WellKnownMediaType && offered != gin.MIMEJSON {
		reportProblem(c,
			http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output format is %s",
				capability.WellKnownMediaType),
		)
		return
	}

	obj, err := capability.NewWellKnownInfoObj(
		nil, // key
		[]string{EdApiMediaType},
		nil, // supported schemes
		config.Version,
		"SERVICE_STATUS_READY",
		publicApiMap,
	)

	if err != nil {
		reportProblem(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Header("Content-Type", capability.WellKnownMediaType)
	c.JSON(http.StatusOK, obj)
}

func (o Handler) respondToGet(c *gin.Context, mt string, ret interface{}, err error) {
	if err != nil {
		status := http.StatusBadRequest

		if errors.Is(err, errTodo) {
			status = http.StatusInternalServerError
		}

		reportProblem(c, status, err.Error())
		return
	}

	respBytes, err := json.Marshal(ret)
	if err != nil {
		reportProblem(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Data(http.StatusOK, mt, respBytes)
}

func reportProblem(c *gin.Context, status int, details ...string) {
	prob := problems.NewStatusProblem(status)

	if len(details) > 0 {
		prob.Detail = strings.Join(details, ", ")
	}

	log.LogProblem(log.Named("api"), prob)

	c.Header("Content-Type", "application/problem+json")
	c.AbortWithStatusJSON(status, prob)
}

func (o Handler) CoservRequest(c *gin.Context) {
	offered := c.NegotiateFormat(EdApiMediaType)
	if offered != EdApiMediaType {
		reportProblem(c,
			http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output format is %s",
				EdApiMediaType),
		)
		return
	}

	coservQuery := c.Param("query")

	var coserv coserv.Coserv
	if err := coserv.FromBase64Url(coservQuery); err != nil {
		reportProblem(c, http.StatusBadRequest, err.Error())
		return
	}

	// Forward query to VTS

	o.respondToGet(c, EdApiMediaType, nil, errTodo)
}
