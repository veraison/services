// Copyright 2025-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package coserv

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/go-cose"
	"github.com/veraison/services/pkg/api/common"
	"github.com/veraison/services/pkg/config"
	"github.com/veraison/services/pkg/log"
	"github.com/veraison/services/pkg/util"
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
	logger                 *zap.SugaredLogger
	endorsementDistributor IEndorsementDistributor
	WkCacheMaxAge          time.Duration
}

func NewHandler(
	endorsementdistributor IEndorsementDistributor,
	logger *zap.SugaredLogger,
	wkCacheMaxAge time.Duration,
) Handler {
	return Handler{
		endorsementDistributor: endorsementdistributor,
		logger:                 logger,
		WkCacheMaxAge:          wkCacheMaxAge,
	}
}

func (o Handler) GetEdApiWellKnownInfo(c *gin.Context) {
	acceptable := []string{CoservDiscoveryMediaType, gin.MIMEJSON}

	if c.NegotiateFormat(acceptable...) == "" {
		common.ReportProblem(o.logger, c,
			http.StatusNotAcceptable,
			fmt.Sprintf("supported format(s): %s",
				strings.Join(acceptable, ", ")),
		)
		return
	}

	vtsState, err := o.endorsementDistributor.ServerState()
	if err != nil {
		o.logger.Error(err)
		common.ReportProblem(o.logger, c,
			http.StatusInternalServerError,
			"Internal Server Error",
		)
		return
	}

	// Get public key
	key, err := util.ParseJWK(vtsState.CoservSigningPublicKey)
	if err != nil {
		o.logger.Error(err)
		common.ReportProblem(o.logger, c,
			http.StatusInternalServerError,
			"Internal Server Error",
		)
		return
	}

	capabilities := make([]Capability, 0, len(vtsState.CoservMediaTypes)) // preallocate

	for _, mt := range vtsState.CoservMediaTypes {
		c := Capability{
			MediaType:       mt,
			ArtifactSupport: []string{"collected"}, // only "collected" supported for now
		}
		capabilities = append(capabilities, c)
	}

	obj := NewCoservWellKnownInfo(
		config.Version,
		capabilities,
		publicApiMap,
		[]jwk.Key{key},
	)

	c.Header("Cache-Control", fmt.Sprintf("max-age=%d", int64(o.WkCacheMaxAge.Seconds())))
	c.Header("Expires", time.Now().Add(o.WkCacheMaxAge).UTC().Format(time.RFC1123))
	c.Header("Content-Type", CoservDiscoveryMediaType)
	c.JSON(http.StatusOK, obj)
}

func (o Handler) CoservRequest(c *gin.Context) {
	offered := c.NegotiateFormat(CoservMTs...)
	if !slices.Contains(CoservMTs, offered) {
		common.ReportConciseProblem(o.logger, c,
			http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output formats are %s",
				strings.Join(CoservMTs, " or ")),
		)
		return
	}

	coservQuery := c.Param("query")

	var coservData coserv.Coserv
	if err := coservData.FromBase64Url(coservQuery); err != nil {
		common.ReportConciseProblem(o.logger, c, http.StatusBadRequest, err.Error())
		return
	}

	profile, err := coservData.Profile.Get()
	if err != nil {
		common.ReportConciseProblem(o.logger, c, http.StatusBadRequest, err.Error())
		return
	}
	mediaType := fmt.Sprintf(`%s; profile=%q`, offered, profile)

	// Forward query to VTS
	res, err := o.endorsementDistributor.GetEndorsements(tenantID, coservQuery, mediaType)
	if err != nil {
		status := http.StatusBadRequest
		common.ReportConciseProblem(o.logger, c, status, err.Error())
		return
	}

	// Set caching headers based on the result set expiry timestamp
	setCacheHeaders(c, res)

	c.Data(http.StatusOK, mediaType, res)
}

// unwrapCoserv attempts to parse the input bytes as a COSE_Sign1 message and
// returns the payload if successful; otherwise, it returns the input bytes
// unchanged. This allows setCacheHeaders to extract the CoSERV result set from
// either a signed or unsigned message.
func unwrapCoserv(data []byte) []byte {
	var msg cose.Sign1Message

	if err := msg.UnmarshalCBOR(data); err == nil {
		return msg.Payload
	}

	return data
}

// setCacheHeaders sets HTTP caching headers based on the expiry timestamp in
// the CoSERV result set.
func setCacheHeaders(c *gin.Context, resultBytes []byte) {
	var result coserv.Coserv

	resultBytes = unwrapCoserv(resultBytes)

	// §6.1.3 of draft-ietf-rats-coserv: "the HTTP cache directives [...] MUST
	// NOT exceed the result set expiry timestamp.

	if err := result.FromCBOR(resultBytes); err != nil {
		log.Warnw("failed to unmarshal result", "error", err)
		return
	}

	if result.Results == nil || result.Results.Expiry == nil {
		// XXX this should not happen as VTS is expected to always include an
		// expiry timestamp in the result set, but log a warning and skip
		// setting cache headers if it does
		log.Warn("ResultSet has no expiry timestamp, skipping cache headers")
		return
	}

	exp := result.Results.Expiry

	// XXX the following assumes VTS and CoSERV are time-synchronized
	now := time.Now()
	maxAge := int64(exp.Sub(now).Seconds())

	// Ensure max-age is non-negative; if expiry is in the past, use 0
	if maxAge < 0 {
		maxAge = 0
	}

	c.Header("Cache-Control", fmt.Sprintf("max-age=%d", maxAge))
	c.Header("Expires", exp.UTC().Format(time.RFC1123))
}
