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
	acceptable := []string{CoservDiscoveryMediaType, gin.MIMEJSON}

	// TODO (tho) - add reportCBORProblem and use it here
	if c.NegotiateFormat(acceptable...) == "" {
		reportProblem(c,
			http.StatusNotAcceptable,
			fmt.Sprintf("supported format(s): %s",
				strings.Join(acceptable, ", ")),
		)
		return
	}

	profiles, err := o.EndorsementDistibutor.SupportedProfiles()
	if err != nil {
		reportProblem(c, http.StatusInternalServerError, err.Error())
		return
	}

	var capabilities []Capability

	for _, p := range profiles {
		c := Capability{
			MediaType:       fmt.Sprint(EdApiMediaType, `; profile="`, p, `"`),
			ArtifactSupport: []string{"collected"}, // only "collected" supported for now
		}
		capabilities = append(capabilities, c)
	}

	// TODO(tho)
	// - capabilities
	// - keys (reuse EAR Signer?)
	obj := NewCoservWellKnownInfo(
		config.Version,
		capabilities,
		publicApiMap,
		nil, // TODO keys
	)

	c.Header("Content-Type", CoservDiscoveryMediaType)
	c.JSON(http.StatusOK, obj)
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
