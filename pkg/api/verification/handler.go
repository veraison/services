// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package verification

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/veraison/cmw"
	"github.com/veraison/services/pkg/api/common"
	"github.com/veraison/services/pkg/api/verification/session"
	"github.com/veraison/services/pkg/fault"
	"github.com/veraison/services/pkg/log"
	"github.com/veraison/services/pkg/util"
	"go.uber.org/zap"
)

const (
	ChallengeResponseSessionMediaType = "application/vnd.veraison.challenge-response-session+json"
)

var (
	ErrInternal = errors.New("internal error")

	tenantID = "0"
)

type IHandler interface {
	NewChallengeResponse(c *gin.Context)
	SubmitEvidence(c *gin.Context)
	GetSession(c *gin.Context)
	DelSession(c *gin.Context)
	GetWellKnownVerificationInfo(c *gin.Context)
}

type Handler struct {
	SessionManager session.IManager
	Verifier       IVerifier
	WkCacheMaxAge  time.Duration
	MaxPayloadSize int64
	logger         *zap.SugaredLogger
}

func NewHandler(
	sessionManager session.IManager,
	verifier IVerifier,
	wkCacheMaxAge time.Duration,
	maxPayloadSize int64,
) IHandler {
	logger := log.Named("api-handler")

	return &Handler{
		SessionManager: sessionManager,
		Verifier:       verifier,
		WkCacheMaxAge:  wkCacheMaxAge,
		MaxPayloadSize: maxPayloadSize,
		logger:         logger,
	}
}

func (o *Handler) GetSession(c *gin.Context) {
	offered := c.NegotiateFormat(ChallengeResponseSessionMediaType)
	if offered != ChallengeResponseSessionMediaType {
		common.ReportProblem(o.logger, c,
			http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output format is %s", ChallengeResponseSessionMediaType),
		)
		return
	}

	id, err := readSessionIDFromRequestURI(c)
	if err != nil {
		common.ReportProblem(o.logger, c,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	sess, err := lookupSession(o.SessionManager, id, tenantID)
	if err != nil {
		common.ReportProblem(o.logger, c,
			http.StatusNotFound,
			err.Error(),
		)
		return
	}

	c.Header("Content-Type", ChallengeResponseSessionMediaType)
	c.JSON(http.StatusOK, sess)
}

func (o *Handler) DelSession(c *gin.Context) {
	id, err := readSessionIDFromRequestURI(c)
	if err != nil {
		common.ReportProblem(o.logger, c,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	if err = o.SessionManager.DelSession(id, tenantID); err != nil {
		common.ReportProblem(o.logger, c,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	c.Status(http.StatusNoContent)
}

func (o *Handler) SubmitEvidence(c *gin.Context) {
	// do content negotiation (accept application/vnd.veraison.challenge-response-session+json)
	offered := c.NegotiateFormat(ChallengeResponseSessionMediaType)
	if offered != ChallengeResponseSessionMediaType {
		common.ReportProblem(o.logger, c,
			http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output format is %s", ChallengeResponseSessionMediaType),
		)
		return
	}

	// read body (i.e., evidence)
	evidence, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, o.MaxPayloadSize))
	if err != nil || len(evidence) == 0 {
		o.logger.Error("unable to read evidence from the request body: %v", err)

		msg := "unable to read evidence from the request body"
		if err != nil && err.Error() == "http: request body too large" {
			msg = "payload too large"
		}

		common.ReportProblem(o.logger, c,
			http.StatusBadRequest,
			msg,
		)
		return
	}

	// read content-type and check against supported attestation formats
	mediaType := c.Request.Header.Get("Content-Type")

	if isCMW(mediaType) {
		var w cmw.CMW

		if err := w.Deserialize(evidence); err != nil {
			common.ReportProblem(o.logger, c,
				http.StatusBadRequest,
				fmt.Sprintf("could not unwrap the CMW: %v", err),
			)
			return
		}

		// Here we only deal with CMW records and tags.  Collection CMWs are
		// dealt with in the non-exceptional path.
		if w.GetKind() == cmw.KindMonad {
			mediaType, err = w.GetMonadType()
			if err != nil {
				common.ReportProblem(o.logger, c,
					http.StatusBadRequest,
					fmt.Sprintf("could not extract CMW media type: %v", err),
				)
			}

			evidence, err = w.GetMonadValue()
			if err != nil {
				common.ReportProblem(o.logger, c,
					http.StatusBadRequest,
					fmt.Sprintf("could not extract CMW value: %v", err),
				)
			}
		}
	}

	isSupported, err := o.Verifier.IsSupportedMediaType(mediaType)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, fault.ErrBadInput) {
			status = http.StatusBadRequest
		}

		common.ReportProblem(o.logger, c,
			status,
			fmt.Sprintf("could not check media type with verifier: %v", err))
		return
	}

	if !isSupported {
		supportedMediaTypes, err := o.Verifier.SupportedMediaTypes()
		if err != nil {
			common.ReportProblem(o.logger, c,
				http.StatusInternalServerError,
				fmt.Sprintf("could not get supported media types from verifier: %v",
					err),
			)
			return
		}

		c.Header("Accept", strings.Join(supportedMediaTypes, ", "))
		common.ReportProblem(o.logger, c,
			http.StatusUnsupportedMediaType,
			fmt.Sprintf("no active plugin found for %s", mediaType),
		)
		return
	}

	id, err := readSessionIDFromRequestURI(c)
	if err != nil {
		common.ReportProblem(o.logger, c,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	// load currentSession from request URI
	currentSession, err := lookupSession(o.SessionManager, id, tenantID)
	if err != nil {
		common.ReportProblem(o.logger, c,
			http.StatusNotFound,
			err.Error(),
		)
		return
	}

	// Forward the evidence to the verifier. We expect the verifier to be
	// able to cope with bad evidence, so the error here should only be
	// reported if something in the verifier or the connection goes wrong.
	// Any problems with the evidence are expected to be reported via the
	// attestation result.
	attestationResult, err := o.Verifier.ProcessEvidence(tenantID, currentSession.Nonce,
		evidence, mediaType)
	if err != nil {
		o.logger.Error(err)
		currentSession.SetStatus(session.StatusFailed)
		mustStoreSession(o.SessionManager, currentSession, id, tenantID)
		common.ReportProblem(o.logger, c,
			http.StatusInternalServerError,
			"error encountered while processing evidence",
		)
		return
	}

	// From here onwards, any signalling to the client (including failure
	// is done through the session object.

	currentSession.SetEvidence(mediaType, evidence)

	// async (202)
	if attestationResult == nil {
		currentSession.SetStatus(session.StatusProcessing)
		s := mustStoreSession(o.SessionManager, currentSession, id, tenantID)
		sendChallengeResponseSessionWithStatus(c, http.StatusAccepted, s)
		return
	}

	// sync (200)
	currentSession.SetStatus(session.StatusComplete)
	currentSession.SetResult(attestationResult)
	s := mustStoreSession(o.SessionManager, currentSession, id, tenantID)
	sendChallengeResponseSessionWithStatus(c, http.StatusOK, s)
}

func (o *Handler) NewChallengeResponse(c *gin.Context) {
	offered := c.NegotiateFormat(ChallengeResponseSessionMediaType)
	if offered != ChallengeResponseSessionMediaType {
		common.ReportProblem(o.logger, c,
			http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output format is %s", ChallengeResponseSessionMediaType),
		)
		return
	}

	// parse query to devise the nonce
	sessionNonce, err := parseNonceRequest(c.Query("nonce"), c.Query("nonceSize"))
	if err != nil {
		status := http.StatusBadRequest

		if errors.Is(err, ErrInternal) {
			status = http.StatusInternalServerError
		}

		common.ReportProblem(o.logger, c,
			status,
			fmt.Sprintf("failed handling nonce request: %s", err),
		)
		return
	}

	supportedMediaTypes, err := o.Verifier.SupportedMediaTypes()
	if err != nil {
		common.ReportProblem(o.logger, c,
			http.StatusInternalServerError,
			fmt.Sprintf("could not get media types form verifier: %v", err),
		)
		return
	}

	id, sess, err := newSession(sessionNonce, supportedMediaTypes, ConfigSessionTTL)
	if err != nil {
		common.ReportProblem(o.logger, c,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	err = o.SessionManager.SetSession(id, tenantID, sess, ConfigSessionTTL)
	if err != nil {
		common.ReportProblem(o.logger, c,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	sendChallengeResponseSessionCreated(c, id.String(), sess)
}

var cmwMap = map[string]bool{
	"application/cmw":                   true,
	"application/cmw+json":              true,
	"application/cmw+cbor":              true,
	"application/vnd.veraison.cmw":      true,
	"application/vnd.veraison.cmw+cbor": true,
	"application/vnd.veraison.cmw+json": true,
}

func isCMW(mt string) bool {
	_, ok := cmwMap[mt]
	return ok
}

func sendChallengeResponseSessionWithStatus(c *gin.Context, status int, jsonSession []byte) {
	c.Data(status, ChallengeResponseSessionMediaType, jsonSession)
}

func sendChallengeResponseSessionCreated(c *gin.Context, id string, jsonSession []byte) {
	c.Header("Location", path.Join("session", id))
	sendChallengeResponseSessionWithStatus(c, http.StatusCreated, jsonSession)
}

func getVerificationEndpoints() map[string]string {
	return publicApiMap
}

func (o *Handler) GetWellKnownVerificationInfo(c *gin.Context) {
	offered := c.NegotiateFormat(common.WellKnownMediaType)
	if offered != common.WellKnownMediaType && offered != gin.MIMEJSON {
		common.ReportProblem(o.logger, c,
			http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output format is %s", common.WellKnownMediaType),
		)
		return
	}

	vtsState, err := o.Verifier.ServerState()
	if err != nil {
		o.logger.Error(err)
		common.ReportProblem(o.logger, c,
			http.StatusInternalServerError,
			"Internal Server Error",
		)
		return
	}

	// Get public key
	key, err := util.ParseJWK(vtsState.EarSigningPublicKey)
	if err != nil {
		o.logger.Error(err)
		common.ReportProblem(o.logger, c,
			http.StatusInternalServerError,
			"Internal Server Error",
		)
		return
	}

	// Get verification endpoints
	endpoints := getVerificationEndpoints()

	// Get final object with well known information
	obj, err := common.NewWellKnownInfoObj(
		key,
		vtsState.VerificationMediaTypes,
		nil,
		vtsState.Version,
		endpoints,
	)
	if err != nil {
		o.logger.Error(err)
		common.ReportProblem(o.logger, c,
			http.StatusInternalServerError,
			"Internal Server Error",
		)
		return
	}

	c.Header("Cache-Control", fmt.Sprintf("max-age=%d", int64(o.WkCacheMaxAge.Seconds())))
	c.Header("Expires", time.Now().Add(o.WkCacheMaxAge).UTC().Format(time.RFC1123))
	c.Header("Content-Type", common.WellKnownMediaType)
	c.JSON(http.StatusOK, obj)
}

var (
	ConfigNonceSize     uint8 = 32
	ConfigSessionTTL, _       = time.ParseDuration("2m30s")
)

// mintSessionID creates a version 1 UUID based on a unique machine ID, clock
// sequence and current time.  Routing to the correct node can therefore happen
// based on the NodeID part of the UUID (i.e., octets 10-15).
func mintSessionID() (uuid.UUID, error) {
	mid, err := machineid.ID()
	if err != nil {
		return uuid.UUID{}, err
	}

	uuid.SetNodeID([]byte(mid))

	return uuid.NewUUID()
}

// mintNonce creates a random nonce of nonceSz bytes.  nonceSz must be strictly
// positive
func mintNonce(nonceSz uint8) ([]byte, error) {
	if nonceSz == 0 {
		return nil, errors.New("nonce size cannot be 0")
	}

	n := make([]byte, nonceSz)

	_, err := rand.Read(n)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternal, err)
	}

	return n, nil
}

// aToU8 attempts at converting the supplied string into an uint8 value
func aToU8(v string) (uint8, error) {
	u8, err := strconv.ParseUint(v, 10, 8)
	if err != nil {
		return ^uint8(0), err
	}

	return uint8(u8), nil
}

// parseNonceRequest tries to devise the nonce value to be used for the session
// given the user-supplied query parameters
func parseNonceRequest(nonceParam, nonceSizeParam string) ([]byte, error) {
	// both nonce and nonceSize have been supplied
	if nonceParam != "" && nonceSizeParam != "" {
		return nil, errors.New("nonce and nonceSize are mutually exclusive")
	}

	// no explicit request was made, use the default nonce size to mint a
	// new nonce
	if nonceParam == "" && nonceSizeParam == "" {
		return mintNonce(ConfigNonceSize)
	}

	// a nonceSize was supplied, try to use it to mint a new nonce
	if nonceSizeParam != "" {
		nonceSize, err := aToU8(nonceSizeParam)
		if err != nil || nonceSize < 8 || nonceSize > 64 {
			return nil, errors.New("nonceSize must be in range 8..64")
		}
		return mintNonce(nonceSize)
	}

	// nonce was supplied, try to see if the encoding is valid
	nonce, err := util.Base64URLToBytes(nonceParam)
	if err != nil {
		return nil, errors.New("nonce must be valid base64")
	}

	nonceLen := len(nonce)
	if nonceLen < 8 || nonceLen > 64 {
		return nil, fmt.Errorf(
			"nonce must be between 8 and 64 bytes long; got %d",
			nonceLen,
		)
	}

	return nonce, nil
}

func newSession(sessionNonce []byte, supportedMediaTypes []string, ttl time.Duration) (uuid.UUID, []byte, error) {
	id, err := mintSessionID()
	if err != nil {
		return uuid.UUID{}, nil, err
	}

	sess := &session.Session{
		ID:     id.String(),
		Status: session.StatusWaiting, // start in waiting status
		Nonce:  sessionNonce,
		Expiry: time.Now().Add(ttl), // RFC3339 format, with sub-second precision added if present
		Accept: supportedMediaTypes,
	}

	jsonSession, err := json.Marshal(sess)
	if err != nil {
		return uuid.UUID{}, nil, err
	}

	return id, jsonSession, nil
}

func lookupSession(sm session.IManager, id uuid.UUID, tenantID string) (*session.Session, error) {
	encodedSession, err := sm.GetSession(id, tenantID)
	if err != nil {
		return nil, err
	}

	var result session.Session
	err = json.Unmarshal(encodedSession, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func storeSession(sm session.IManager, sess *session.Session, id uuid.UUID, tenantID string) ([]byte, error) {
	b, err := json.Marshal(sess)
	if err != nil {
		return nil, err
	}

	err = sm.SetSession(id, tenantID, b, ConfigSessionTTL)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func mustStoreSession(sm session.IManager, sess *session.Session, id uuid.UUID, tenantID string) []byte {
	s, err := storeSession(sm, sess, id, tenantID)
	if err != nil {
		panic(err)
	}

	return s
}

func readSessionIDFromRequestURI(c *gin.Context) (uuid.UUID, error) {
	uriPathSegment := c.Param("id")

	id, err := uuid.Parse(uriPathSegment)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf(
			"invalid session id (%s) in path segment: %w",
			c.Request.URL.Path,
			err,
		)
	}

	return id, nil
}
