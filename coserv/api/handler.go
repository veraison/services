// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/moogar0880/problems"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/services/capability"
	"github.com/veraison/services/config"
	"github.com/veraison/services/coserv/endorsementdistributor"
	"github.com/veraison/services/log"
	"go.uber.org/zap"
)

var (
	tenantID       = "0"
	EdApiMediaType = "application/coserv+cbor"
	errTodo        = errors.New("TODO")
)

type Handler struct {
	Logger                *zap.SugaredLogger
	EndorsementDistibutor endorsementdistributor.IEndorsementDistributor
}

func NewHandler(endorsementdistributor endorsementdistributor.IEndorsementDistributor, logger *zap.SugaredLogger) Handler {
	return Handler{
		EndorsementDistibutor: endorsementdistributor,
		Logger:                logger,
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

	// TODO(tho)
	// - supported media types
	// - vts state

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

func (o Handler) respondToGet(c *gin.Context, mt string, ret []byte, err error) {
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

	profile, err := coserv.Profile.Get()
	if err != nil {
		reportProblem(c, http.StatusBadRequest, err.Error())
		return
	}
	mediaType := fmt.Sprintf(`%s; profile=%q`, offered, profile)

	// Forward query to VTS
	res, err := o.EndorsementDistibutor.GetEndorsements(tenantID, coservQuery, mediaType)
	if err != nil {
		status := http.StatusBadRequest

		if errors.Is(err, errTodo) {
			status = http.StatusInternalServerError
		}

		reportProblem(c, status, err.Error())
		return
	}

	c.Data(http.StatusOK, EdApiMediaType, res)
}
