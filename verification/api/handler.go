// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package api

import (
	"crypto/rand"
	"encoding/base64"
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
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/veraison/cmw"
	"github.com/veraison/services/capability"
	"github.com/veraison/services/log"
	"github.com/veraison/services/verification/sessionmanager"
	"github.com/veraison/services/verification/verifier"
	"go.uber.org/zap"
)

const (
	ChallengeResponseSessionMediaType = "application/vnd.veraison.challenge-response-session+json"
)

var (
	ErrInternal = errors.New("internal error")
)

var (
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
	SessionManager sessionmanager.ISessionManager
	Verifier       verifier.IVerifier

	logger *zap.SugaredLogger
}

func NewHandler(sm sessionmanager.ISessionManager, v verifier.IVerifier) IHandler {
	return &Handler{
		SessionManager: sm,
		Verifier:       v,
		logger:         log.Named("api-handler"),
	}
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

// b64ToBytes attempts at converting the supplied b64-encoded string into a byte
// slice
func b64ToBytes(v string) ([]byte, error) {
	b, err := base64.URLEncoding.DecodeString(v)
	if err != nil {
		return nil, err
	}

	return b, nil
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
	nonce, err := b64ToBytes(nonceParam)
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

func newSession(nonce []byte, supportedMediaTypes []string, ttl time.Duration) (uuid.UUID, []byte, error) {
	id, err := mintSessionID()
	if err != nil {
		return uuid.UUID{}, nil, err
	}

	session := &ChallengeResponseSession{
		id:     id.String(),
		Status: StatusWaiting, // start in waiting status
		Nonce:  nonce,
		Expiry: time.Now().Add(ttl), // RFC3339 format, with sub-second precision added if present
		Accept: supportedMediaTypes,
	}

	jsonSession, err := json.Marshal(session)
	if err != nil {
		return uuid.UUID{}, nil, err
	}

	return id, jsonSession, nil
}

func lookupSession(sm sessionmanager.ISessionManager, id uuid.UUID, tenantID string) (*ChallengeResponseSession, error) {
	session, err := sm.GetSession(id, tenantID)
	if err != nil {
		return nil, err
	}

	var s ChallengeResponseSession

	err = json.Unmarshal(session, &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func storeSession(sm sessionmanager.ISessionManager, session *ChallengeResponseSession, id uuid.UUID, tenantID string) ([]byte, error) {
	b, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}

	err = sm.SetSession(id, tenantID, b, ConfigSessionTTL)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func mustStoreSession(sm sessionmanager.ISessionManager, session *ChallengeResponseSession, id uuid.UUID, tenantID string) []byte {
	s, err := storeSession(sm, session, id, tenantID)
	if err != nil {
		panic(err)
	}

	return s
}

func readSessionIDFromRequestURI(c *gin.Context) (uuid.UUID, error) {
	uriPathSegment := c.Param("id")

	id, err := uuid.Parse(uriPathSegment)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid session id (%s) in path segment: %w", c.Request.URL.Path, err)
	}

	return id, nil
}

func (o *Handler) GetSession(c *gin.Context) {
	// do content negotiation (accept application/vnd.veraison.challenge-response-session+json)
	offered := c.NegotiateFormat(ChallengeResponseSessionMediaType)
	if offered != ChallengeResponseSessionMediaType {
		ReportProblem(c,
			http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output format is %s", ChallengeResponseSessionMediaType),
		)
		return
	}

	id, err := readSessionIDFromRequestURI(c)
	if err != nil {
		ReportProblem(c,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	// load session from request URI
	session, err := lookupSession(o.SessionManager, id, tenantID)
	if err != nil {
		ReportProblem(c,
			http.StatusNotFound,
			err.Error(),
		)
		return
	}

	c.Header("Content-Type", ChallengeResponseSessionMediaType)
	c.JSON(http.StatusOK, session)
}

func (o *Handler) DelSession(c *gin.Context) {
	id, err := readSessionIDFromRequestURI(c)
	if err != nil {
		ReportProblem(c,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	if err = o.SessionManager.DelSession(id, tenantID); err != nil {
		ReportProblem(c,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	c.Status(http.StatusNoContent)
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

func (o *Handler) SubmitEvidence(c *gin.Context) {
	// do content negotiation (accept application/vnd.veraison.challenge-response-session+json)
	offered := c.NegotiateFormat(ChallengeResponseSessionMediaType)
	if offered != ChallengeResponseSessionMediaType {
		ReportProblem(c,
			http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output format is %s", ChallengeResponseSessionMediaType),
		)
		return
	}

	// read body (i.e., evidence)
	evidence, err := io.ReadAll(c.Request.Body)
	if err != nil || len(evidence) == 0 {
		o.logger.Error("unable to read evidence from the request body: %v", err)
		ReportProblem(c,
			http.StatusBadRequest,
			"unable to read evidence from the request body",
		)
		return
	}

	// read content-type and check against supported attestation formats
	mediaType := c.Request.Header.Get("Content-Type")

	if isCMW(mediaType) {
		var w cmw.CMW

		if err := w.Deserialize(evidence); err != nil {
			ReportProblem(c,
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
				ReportProblem(c,
					http.StatusBadRequest,
					fmt.Sprintf("could not extract CMW media type: %v", err),
				)
			}

			evidence, err = w.GetMonadValue()
			if err != nil {
				ReportProblem(c,
					http.StatusBadRequest,
					fmt.Sprintf("could not extract CMW value: %v", err),
				)
			}
		}
	}

	isSupported, err := o.Verifier.IsSupportedMediaType(mediaType)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Unwrap(err) == verifier.ErrInputParam {
			status = http.StatusBadRequest
		}

		ReportProblem(c, status, fmt.Sprintf("could not check media type with verifier: %v", err))
		return
	}

	if !isSupported {
		supportedMediaTypes, err := o.Verifier.SupportedMediaTypes()
		if err != nil {
			ReportProblem(c,
				http.StatusInternalServerError,
				fmt.Sprintf("could not get supported media types from verifier: %v",
					err),
			)
			return
		}

		c.Header("Accept", strings.Join(supportedMediaTypes, ", "))
		ReportProblem(c,
			http.StatusUnsupportedMediaType,
			fmt.Sprintf("no active plugin found for %s", mediaType),
		)
		return
	}

	id, err := readSessionIDFromRequestURI(c)
	if err != nil {
		ReportProblem(c,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	// load session from request URI
	session, err := lookupSession(o.SessionManager, id, tenantID)
	if err != nil {
		ReportProblem(c,
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
	attestationResult, err := o.Verifier.ProcessEvidence(tenantID, session.Nonce,
		evidence, mediaType)
	if err != nil {
		o.logger.Error(err)
		session.SetStatus(StatusFailed)
		mustStoreSession(o.SessionManager, session, id, tenantID)
		ReportProblem(c,
			http.StatusInternalServerError,
			"error encountered while processing evidence",
		)
		return
	}

	// From here onwards, any signalling to the client (including failure
	// is done through the session object.

	session.SetEvidence(mediaType, evidence)

	// async (202)
	if attestationResult == nil {
		session.SetStatus(StatusProcessing)
		s := mustStoreSession(o.SessionManager, session, id, tenantID)
		sendChallengeResponseSessionWithStatus(c, http.StatusAccepted, s)
		return
	}

	// sync (200)
	session.SetStatus(StatusComplete)
	session.SetResult(attestationResult)
	s := mustStoreSession(o.SessionManager, session, id, tenantID)
	sendChallengeResponseSessionWithStatus(c, http.StatusOK, s)
}

func (o *Handler) NewChallengeResponse(c *gin.Context) {
	offered := c.NegotiateFormat(ChallengeResponseSessionMediaType)
	if offered != ChallengeResponseSessionMediaType {
		ReportProblem(c,
			http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output format is %s", ChallengeResponseSessionMediaType),
		)
		return
	}

	// parse query to devise the nonce
	nonce, err := parseNonceRequest(c.Query("nonce"), c.Query("nonceSize"))
	if err != nil {
		status := http.StatusBadRequest

		if errors.Is(err, ErrInternal) {
			status = http.StatusInternalServerError
		}

		ReportProblem(c,
			status,
			fmt.Sprintf("failed handling nonce request: %s", err),
		)
		return
	}

	supportedMediaTypes, err := o.Verifier.SupportedMediaTypes()
	if err != nil {
		ReportProblem(c,
			http.StatusInternalServerError,
			fmt.Sprintf("could not get media types form verifier: %v", err),
		)
		return
	}

	id, session, err := newSession(nonce, supportedMediaTypes, ConfigSessionTTL)
	if err != nil {
		ReportProblem(c,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	err = o.SessionManager.SetSession(id, tenantID, session, ConfigSessionTTL)
	if err != nil {
		ReportProblem(c,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	sendChallengeResponseSessionCreated(c, id.String(), session)
}

func sendChallengeResponseSessionWithStatus(c *gin.Context, status int, jsonSession []byte) {
	c.Data(status, ChallengeResponseSessionMediaType, jsonSession)
}

func sendChallengeResponseSessionCreated(c *gin.Context, id string, jsonSession []byte) {
	c.Header("Location", path.Join("session", id))
	sendChallengeResponseSessionWithStatus(c, http.StatusCreated, jsonSession)
}

func (o *Handler) getKey() (jwk.Key, error) {
	var key jwk.Key

	protoKey, err := o.Verifier.GetPublicKey()
	if err != nil {
		return key, err
	}

	key, err = jwk.ParseKey([]byte(protoKey.Key))
	if err != nil {
		return key, err
	}

	return key, nil

}

func (o *Handler) getVerificationMediaTypes() ([]string, error) {
	return o.Verifier.SupportedMediaTypes()
}

func (o *Handler) getVerificationServerVersionAndState() (string, string, error) {
	vtsState, err := o.Verifier.GetVTSState()
	if err != nil {
		return "", "", err
	}
	version := vtsState.ServerVersion
	state := vtsState.Status.String()
	return version, state, nil

}

func getVerificationEndpoints() map[string]string {
	return publicApiMap
}

func (o *Handler) GetWellKnownVerificationInfo(c *gin.Context) {
	offered := c.NegotiateFormat(capability.WellKnownMediaType)
	if offered != capability.WellKnownMediaType && offered != gin.MIMEJSON {
		ReportProblem(c,
			http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output format is %s", capability.WellKnownMediaType),
		)
		return
	}

	// Get public key
	key, err := o.getKey()
	if err != nil {
		ReportProblem(c,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	// Get verification media types
	mediaTypes, err := o.getVerificationMediaTypes()
	if err != nil {
		ReportProblem(c,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	// Get verification server version and state
	version, state, err := o.getVerificationServerVersionAndState()
	if err != nil {
		ReportProblem(c,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	// Get verification endpoints
	endpoints := getVerificationEndpoints()

	// Get final object with well known information
	obj, err := capability.NewWellKnownInfoObj(key, mediaTypes, nil, version, state, endpoints)
	if err != nil {
		ReportProblem(c,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	c.Header("Content-Type", capability.WellKnownMediaType)
	c.JSON(http.StatusOK, obj)
}
