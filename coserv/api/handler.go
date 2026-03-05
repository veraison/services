// Copyright 2025-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/go-cose"
	"github.com/veraison/services/capability"
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
	defaultCacheMaxAge = 3600 * time.Second
)

type Handler struct {
	Logger                *zap.SugaredLogger
	EndorsementDistibutor endorsementdistributor.IEndorsementDistributor
	WkCacheMaxAge         time.Duration
}

func NewHandler(
	endorsementdistributor endorsementdistributor.IEndorsementDistributor,
	logger *zap.SugaredLogger,
	wkCacheMaxAge string,
) Handler {
	return Handler{
		EndorsementDistibutor: endorsementdistributor,
		Logger:                logger,
		WkCacheMaxAge:         capability.ParseCacheMaxAge(wkCacheMaxAge, defaultCacheMaxAge, logger),
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

	c.Header("Cache-Control", fmt.Sprintf("max-age=%d", int64(o.WkCacheMaxAge.Seconds())))
	c.Header("Expires", time.Now().Add(o.WkCacheMaxAge).UTC().Format(time.RFC1123))
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

	// Set caching headers based on the result set expiry timestamp
	setCacheHeaders(c, res)

	c.Data(http.StatusOK, mediaType, res)
}
