// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/fxamacker/cbor/v2"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/services/config"
	"github.com/veraison/services/coserv/endorsementdistributor"
	"github.com/veraison/services/log"
	"github.com/veraison/services/proto"
	"go.uber.org/zap"
)

var (
	tenantID  = "0"
	CoservMTs = []string{
		"application/coserv+cbor",
		"application/coserv+cose",
	}
)

type Handler struct {
	Logger                *zap.SugaredLogger
	EndorsementDistibutor endorsementdistributor.IEndorsementDistributor
}

func NewHandler(
	endorsementdistributor endorsementdistributor.IEndorsementDistributor,
	logger *zap.SugaredLogger,
) Handler {
	return Handler{
		EndorsementDistibutor: endorsementdistributor,
		Logger:                logger,
	}
}

func (o Handler) GetEdApiWellKnownInfo(c *gin.Context) {
	acceptable := []string{CoservDiscoveryMediaType, gin.MIMEJSON}

	if c.NegotiateFormat(acceptable...) == "" {
		reportProblem(c,
			http.StatusNotAcceptable,
			fmt.Sprintf("supported format(s): %s",
				strings.Join(acceptable, ", ")),
		)
		return
	}

	mediaTypes, err := o.EndorsementDistibutor.SupportedMediaTypes()
	if err != nil {
		reportProblem(c, http.StatusInternalServerError, err.Error())
		return
	}

	capabilities := make([]Capability, 0, len(mediaTypes)) // preallocate

	for _, mt := range mediaTypes {
		c := Capability{
			MediaType:       mt,
			ArtifactSupport: []string{"collected"}, // only "collected" supported for now
		}
		capabilities = append(capabilities, c)
	}

	k, err := o.EndorsementDistibutor.GetPublicKey()
	if err != nil {
		reportProblem(c, http.StatusInternalServerError, err.Error())
		return
	}

	keySet, err := toJWKKeySet(k)
	if err != nil {
		reportProblem(c, http.StatusInternalServerError, err.Error())
		return
	}

	obj := NewCoservWellKnownInfo(
		config.Version,
		capabilities,
		publicApiMap,
		keySet,
	)

	c.Header("Content-Type", CoservDiscoveryMediaType)
	c.JSON(http.StatusOK, obj)
}

func toJWKKeySet(k *proto.PublicKey) ([]jwk.Key, error) {
	// if no CoSERV key is configured, return nil (omitempty will skip
	// serialising the discovery object attribute)
	if k == nil || k.Key == "" {
		return nil, nil
	}

	jwkKey, err := jwk.ParseKey([]byte(k.Key))
	if err != nil {
		return nil, fmt.Errorf("parsing public key JWK: %w", err)
	}

	return []jwk.Key{jwkKey}, nil
}

func reportProblem(c *gin.Context, status int, details ...string) {
	type PD struct {
		Title  string `cbor:"-1,keyasint,omitempty"`
		Detail string `cbor:"-2,keyasint,omitempty"`
	}

	prob := PD{
		Title: http.StatusText(status),
	}

	// concatenate details if there are multiple
	if len(details) > 0 {
		prob.Detail = strings.Join(details, ", ")
	}

	log.Info(log.Named("api"), prob)

	b, err := cbor.Marshal(prob)
	if err != nil {
		log.Error(log.Named("api"), "failed to marshal problem details to CBOR", "error", err)
		c.AbortWithStatus(status)
	}

	c.Data(status, "application/concise-problem-details+cbor", b)
}

func (o Handler) CoservRequest(c *gin.Context) {
	offered := c.NegotiateFormat(CoservMTs...)
	if !slices.Contains(CoservMTs, offered) {
		reportProblem(c,
			http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output formats are %s",
				strings.Join(CoservMTs, " or ")),
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
		reportProblem(c, status, err.Error())
		return
	}

	c.Data(http.StatusOK, mediaType, res)
}
