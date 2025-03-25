// Copyright 2022-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/moogar0880/problems"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/cmw"
	"github.com/veraison/services/capability"
	"github.com/veraison/services/proto"
	mock_deps "github.com/veraison/services/verification/api/mocks"
)

const (
	sessionURIRegexp = `^session/[0-9a-fA-F]{8}-([0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}$`
)

var (
	testSupportedMediaTypeA = `application/eat_cwt; profile="http://arm.com/psa/2.0.0"`
	testSupportedMediaTypeB = `application/eat_cwt; profile="PSA_IOT_PROFILE_1"`
	testSupportedMediaTypeC = `application/psa-attestation-token`
	testSupportedMediaTypes = []string{
		testSupportedMediaTypeA,
		testSupportedMediaTypeB,
		testSupportedMediaTypeC,
	}
	testSupportedMediaTypesString = strings.Join(testSupportedMediaTypes, ", ")
	testUnsupportedMediaType      = "application/unknown-evidence-format+json"
	testJSONBody                  = `{ "k": "v" }`
	testSession                   = `{
	"status": "waiting",
	"nonce": "mVubqtg3Wa5GSrx3L/2B99cQU2bMQFVYUI9aTmDYi64=",
	"expiry": "2022-07-13T13:50:24.520525+01:00",
	"accept": [
		"application/eat_cwt;profile=\"http://arm.com/psa/2.0.0\"",
		"application/eat_cwt;profile=\"PSA_IOT_PROFILE_1\"",
		"application/psa-attestation-token"
	]
}`
	testFailedProblem = `{
	"type": "about:blank",
	"title": "Internal Server Error",
	"status": 500,
	"detail": "error encountered while processing evidence"
}`
	testProcessingSession = `{
	"status": "processing",
	"nonce": "mVubqtg3Wa5GSrx3L/2B99cQU2bMQFVYUI9aTmDYi64=",
	"expiry": "2022-07-13T13:50:24.520525+01:00",
	"accept": [
		"application/eat_cwt;profile=\"http://arm.com/psa/2.0.0\"",
		"application/eat_cwt;profile=\"PSA_IOT_PROFILE_1\"",
		"application/psa-attestation-token"
	],
	"evidence": {
		"type":"application/eat_cwt; profile=\"http://arm.com/psa/2.0.0\"",
		"value":"eyAiayI6ICJ2IiB9"
	}
}`
	testCompleteSession = `{
	"status": "complete",
	"nonce": "mVubqtg3Wa5GSrx3L/2B99cQU2bMQFVYUI9aTmDYi64=",
	"expiry": "2022-07-13T13:50:24.520525+01:00",
	"accept": [
		"application/eat_cwt;profile=\"http://arm.com/psa/2.0.0\"",
		"application/eat_cwt;profile=\"PSA_IOT_PROFILE_1\"",
		"application/psa-attestation-token"
	],
	"evidence": {
		"type":"application/eat_cwt; profile=\"http://arm.com/psa/2.0.0\"",
		"value":"eyAiayI6ICJ2IiB9"
	},
	"result": "{}"
}`
	testUUIDString     = "5c5bd88b-c922-482b-ad9f-097e187b42a1"
	testUUID           = uuid.MustParse(testUUIDString)
	testResult         = `{}`
	testNewSessionURL  = "/challenge-response/v1/newSession"
	testSessionBaseURL = "/challenge-response/v1/session"
	testNonce          = []byte{0x99, 0x5b, 0x9b, 0xaa, 0xd8, 0x37, 0x59, 0xae,
		0x46, 0x4a, 0xbc, 0x77, 0x2f, 0xfd, 0x81, 0xf7,
		0xd7, 0x10, 0x53, 0x66, 0xcc, 0x40, 0x55, 0x58,
		0x50, 0x8f, 0x5a, 0x4e, 0x60, 0xd8, 0x8b, 0xae}

	testGoodServiceState = proto.ServiceState{
		Status:        2,
		ServerVersion: "3.2",
	}

	testKeyJSON = `{
		"kty": "EC",
		"alg": "ES256",
		"crv": "P-256",
		"x": "usWxHK2PmfnHKwXPS54m0kTcGJ90UiglWiGahtagnv8",
		"y": "IBOL-C3BttVivg-lSreASjpkttcsz-1rb7btKLv8EX4",
		"d": "V8kgd2ZBRuh2dgyVINBUqpPDr7BOMGcF22CQMIUHtNM"
	}`

	testKey = proto.PublicKey{
		Key: testKeyJSON,
	}
)

func TestHandler_NewChallengeResponse_UnsupportedAccept(t *testing.T) {
	h := &Handler{}

	expectedCode := http.StatusNotAcceptable
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Not Acceptable",
		Status: http.StatusNotAcceptable,
		Detail: fmt.Sprintf("the only supported output format is %s", ChallengeResponseSessionMediaType),
	}

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, "/challenge-response/v1/newSession", http.NoBody)
	req.Header.Set("Accept", "application/unsupported+ber")

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func testHandler_NewChallengeResponse_BadNonce(t *testing.T, queryParams url.Values, expectedErr string) {
	h := &Handler{}

	expectedCode := http.StatusBadRequest
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
		Detail: expectedErr,
	}

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, "/challenge-response/v1/newSession", http.NoBody)
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.URL.RawQuery = queryParams.Encode()

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_NewChallengeResponse_AmbiguousQueryParameters(t *testing.T) {
	q := url.Values{}
	q.Add("nonce", "n")
	q.Add("nonceSize", "1")

	expectedErr := "failed handling nonce request: nonce and nonceSize are mutually exclusive"

	testHandler_NewChallengeResponse_BadNonce(t, q, expectedErr)
}

func TestHandler_NewChallengeResponse_NonceSizeTooBig(t *testing.T) {
	q := url.Values{}
	q.Add("nonceSize", "88")

	expectedErr := "failed handling nonce request: nonceSize must be in range 8..64"

	testHandler_NewChallengeResponse_BadNonce(t, q, expectedErr)
}

func TestHandler_NewChallengeResponse_NonceSizeTooSmall(t *testing.T) {
	q := url.Values{}
	q.Add("nonceSize", "6")

	expectedErr := "failed handling nonce request: nonceSize must be in range 8..64"

	testHandler_NewChallengeResponse_BadNonce(t, q, expectedErr)
}

func TestHandler_NewChallengeResponse_NonceInvalidB64(t *testing.T) {
	q := url.Values{}
	q.Add("nonce", "^^^^")

	expectedErr := "failed handling nonce request: nonce must be valid base64"

	testHandler_NewChallengeResponse_BadNonce(t, q, expectedErr)
}

func TestHandler_NewChallengeResponse_NoNonceParameters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		SetSession(gomock.Any(), tenantID, gomock.Any(), ConfigSessionTTL).
		Return(nil)

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		SupportedMediaTypes().
		Return(testSupportedMediaTypes, nil)

	h := NewHandler(sm, v)

	expectedCode := http.StatusCreated
	expectedType := ChallengeResponseSessionMediaType
	expectedLocationRE := sessionURIRegexp
	expectedSessionStatus := StatusWaiting

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, "/challenge-response/v1/newSession", http.NoBody)
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)

	NewRouter(h).ServeHTTP(w, req)

	var body ChallengeResponseSession
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Regexp(t, expectedLocationRE, w.Result().Header.Get("Location"))
	assert.Len(t, body.Nonce, int(ConfigNonceSize))
	assert.Nil(t, body.Evidence)
	assert.Nil(t, body.Result)
	assert.Equal(t, expectedSessionStatus, body.Status)
}

func TestHandler_NewChallengeResponse_NonceParameter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		SetSession(gomock.Any(), tenantID, gomock.Any(), ConfigSessionTTL).
		Return(nil)

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		SupportedMediaTypes().
		Return(testSupportedMediaTypes, nil)

	h := NewHandler(sm, v)

	expectedCode := http.StatusCreated
	expectedType := ChallengeResponseSessionMediaType
	expectedLocationRE := sessionURIRegexp
	expectedSessionStatus := StatusWaiting
	expectedNonce := []byte("nonce-value")

	qParams := url.Values{}
	// b64("nonce-value") => "bm9uY2UtdmFsdWU="
	qParams.Add("nonce", "bm9uY2UtdmFsdWU=")

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, testNewSessionURL, http.NoBody)
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.URL.RawQuery = qParams.Encode()

	NewRouter(h).ServeHTTP(w, req)

	var body ChallengeResponseSession
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Regexp(t, expectedLocationRE, w.Result().Header.Get("Location"))
	assert.Equal(t, expectedNonce, body.Nonce)
	assert.Nil(t, body.Evidence)
	assert.Nil(t, body.Result)
	assert.Equal(t, expectedSessionStatus, body.Status)
}

func TestHandler_NewChallengeResponse_NonceSizeParameter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		SetSession(gomock.Any(), tenantID, gomock.Any(), ConfigSessionTTL).
		Return(nil)

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		SupportedMediaTypes().
		Return(testSupportedMediaTypes, nil)

	h := NewHandler(sm, v)

	expectedCode := http.StatusCreated
	expectedType := ChallengeResponseSessionMediaType
	expectedLocationRE := sessionURIRegexp
	expectedSessionStatus := StatusWaiting
	expectedNonceSize := 32

	qParams := url.Values{}
	qParams.Add("nonceSize", strconv.Itoa(expectedNonceSize))

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, testNewSessionURL, http.NoBody)
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.URL.RawQuery = qParams.Encode()

	NewRouter(h).ServeHTTP(w, req)

	var body ChallengeResponseSession
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Regexp(t, expectedLocationRE, w.Result().Header.Get("Location"))
	assert.Len(t, body.Nonce, expectedNonceSize)
	assert.Nil(t, body.Evidence)
	assert.Nil(t, body.Result)
	assert.Equal(t, expectedSessionStatus, body.Status)
}

func TestHandler_NewChallengeResponse_SetSessionFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sessionManagerError := "session manager says: doh!"

	expectedCode := http.StatusInternalServerError
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Internal Server Error",
		Status: http.StatusInternalServerError,
		Detail: sessionManagerError,
	}

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		SetSession(gomock.Any(), tenantID, gomock.Any(), ConfigSessionTTL).
		Return(errors.New(sessionManagerError))

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		SupportedMediaTypes().
		Return(testSupportedMediaTypes, nil)

	h := NewHandler(sm, v)

	qParams := url.Values{}
	qParams.Add("nonceSize", "32")

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, testNewSessionURL, http.NoBody)
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.URL.RawQuery = qParams.Encode()

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func testHandler_UnsupportedAccept(t *testing.T, method string) {
	h := &Handler{}

	url := path.Join(testSessionBaseURL, testUUIDString)

	expectedCode := http.StatusNotAcceptable
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Not Acceptable",
		Status: http.StatusNotAcceptable,
		Detail: fmt.Sprintf("the only supported output format is %s", ChallengeResponseSessionMediaType),
	}

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(method, url, http.NoBody)
	req.Header.Set("Accept", "application/unsupported+ber")

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_SubmitEvidence_UnsupportedAccept(t *testing.T) {
	testHandler_UnsupportedAccept(t, http.MethodPost)
}

func TestHandler_SubmitEvidence_unsupported_evidence_format(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	verifierError := "no active plugin found for " + testUnsupportedMediaType

	url := path.Join(testSessionBaseURL, testUUIDString)

	expectedCode := http.StatusUnsupportedMediaType
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Unsupported Media Type",
		Status: http.StatusUnsupportedMediaType,
		Detail: verifierError,
	}

	sm := mock_deps.NewMockISessionManager(ctrl)

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		SupportedMediaTypes().
		Return(testSupportedMediaTypes, nil)
	v.EXPECT().
		IsSupportedMediaType(testUnsupportedMediaType).
		Return(false, nil)

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, url, strings.NewReader(testJSONBody))
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.Header.Set("Content-Type", testUnsupportedMediaType)

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, testSupportedMediaTypesString, w.Result().Header.Get("Accept"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_SubmitEvidence_bad_session_id_url(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	badPath := path.Join(testSessionBaseURL, "1234")

	expectedCode := http.StatusBadRequest
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
		Detail: fmt.Sprintf("invalid session id (%s) in path segment: invalid UUID length: 4", badPath),
	}

	sm := mock_deps.NewMockISessionManager(ctrl)

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		IsSupportedMediaType(testSupportedMediaTypeA).
		Return(true, nil)

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, badPath, strings.NewReader(testJSONBody))
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.Header.Set("Content-Type", testSupportedMediaTypeA)

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_SubmitEvidence_session_not_found(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pathNotFound := path.Join(testSessionBaseURL, testUUIDString)

	smErr := "session not found"

	expectedCode := http.StatusNotFound
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Not Found",
		Status: http.StatusNotFound,
		Detail: smErr,
	}

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		GetSession(testUUID, tenantID).
		Return(nil, errors.New(smErr))

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		IsSupportedMediaType(testSupportedMediaTypeA).
		Return(true, nil)

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, pathNotFound, strings.NewReader(testJSONBody))
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.Header.Set("Content-Type", testSupportedMediaTypeA)

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_SubmitEvidence_no_body(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pathNotFound := path.Join(testSessionBaseURL, testUUIDString)

	expectedCode := http.StatusBadRequest
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
		Detail: "unable to read evidence from the request body",
	}

	sm := mock_deps.NewMockISessionManager(ctrl)
	v := mock_deps.NewMockIVerifier(ctrl)
	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, pathNotFound, http.NoBody)
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.Header.Set("Content-Type", testSupportedMediaTypeA)

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_SubmitEvidence_process_evidence_failed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pathOK := path.Join(testSessionBaseURL, testUUIDString)

	vmErr := "enqueueing evidence failed"

	expectedCode := http.StatusInternalServerError
	expectedType := "application/problem+json"
	expectedBody := testFailedProblem

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		GetSession(testUUID, tenantID).
		Return([]byte(testSession), nil)
	// we cannot assert on the serialised session object (=> gomock.Any()), but
	// it's not a problem because this is going to be checked anyway when
	// matching the response body
	sm.EXPECT().
		SetSession(testUUID, tenantID, gomock.Any(), ConfigSessionTTL).
		Return(nil)

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		IsSupportedMediaType(testSupportedMediaTypeA).
		Return(true, nil)
	v.EXPECT().
		ProcessEvidence(tenantID, testNonce, []byte(testJSONBody), testSupportedMediaTypeA).
		Return(nil, errors.New(vmErr))

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, pathOK, strings.NewReader(testJSONBody))
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.Header.Set("Content-Type", testSupportedMediaTypeA)

	NewRouter(h).ServeHTTP(w, req)

	body := w.Body.Bytes()

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.JSONEq(t, expectedBody, string(body))
}

func TestHandler_SubmitEvidence_process_ok_sync(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pathOK := path.Join(testSessionBaseURL, testUUIDString)

	expectedCode := http.StatusOK
	expectedType := ChallengeResponseSessionMediaType
	expectedBody := testCompleteSession

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		GetSession(testUUID, tenantID).
		Return([]byte(testSession), nil)
	sm.EXPECT().
		SetSession(testUUID, tenantID, gomock.Any(), ConfigSessionTTL).
		Return(nil)

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		IsSupportedMediaType(testSupportedMediaTypeA).
		Return(true, nil)
	v.EXPECT().
		ProcessEvidence(tenantID, testNonce, []byte(testJSONBody), testSupportedMediaTypeA).
		Return([]byte(testResult), nil)

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, pathOK, strings.NewReader(testJSONBody))
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.Header.Set("Content-Type", testSupportedMediaTypeA)

	NewRouter(h).ServeHTTP(w, req)

	body := w.Body.Bytes()

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.JSONEq(t, expectedBody, string(body))
}

func TestHandler_SubmitEvidence_process_ok_async(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pathOK := path.Join(testSessionBaseURL, testUUIDString)

	expectedCode := http.StatusAccepted
	expectedType := ChallengeResponseSessionMediaType
	expectedBody := testProcessingSession

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		GetSession(testUUID, tenantID).
		Return([]byte(testSession), nil)
	sm.EXPECT().
		SetSession(testUUID, tenantID, gomock.Any(), ConfigSessionTTL).
		Return(nil)

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		IsSupportedMediaType(testSupportedMediaTypeA).
		Return(true, nil)
	v.EXPECT().
		ProcessEvidence(tenantID, testNonce, []byte(testJSONBody), testSupportedMediaTypeA).
		Return(nil, nil)

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, pathOK, strings.NewReader(testJSONBody))
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.Header.Set("Content-Type", testSupportedMediaTypeA)

	NewRouter(h).ServeHTTP(w, req)

	body := w.Body.Bytes()

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.JSONEq(t, expectedBody, string(body))
}

func TestHandler_GetSession_UnsupportedAccept(t *testing.T) {
	testHandler_UnsupportedAccept(t, http.MethodGet)
}

func TestHandler_GetSession_bad_session_id_url(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	badPath := path.Join(testSessionBaseURL, "1234")

	expectedCode := http.StatusBadRequest
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
		Detail: fmt.Sprintf("invalid session id (%s) in path segment: invalid UUID length: 4", badPath),
	}

	sm := mock_deps.NewMockISessionManager(ctrl)
	v := mock_deps.NewMockIVerifier(ctrl)

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodGet, badPath, http.NoBody)
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.Header.Set("Content-Type", testSupportedMediaTypeA)

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_GetSession_session_not_found(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pathNotFound := path.Join(testSessionBaseURL, testUUIDString)

	smErr := "session not found"

	expectedCode := http.StatusNotFound
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Not Found",
		Status: http.StatusNotFound,
		Detail: smErr,
	}

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		GetSession(testUUID, tenantID).
		Return(nil, errors.New(smErr))

	v := mock_deps.NewMockIVerifier(ctrl)

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodGet, pathNotFound, http.NoBody)
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.Header.Set("Content-Type", testSupportedMediaTypeA)

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_GetSession_ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pathOK := path.Join(testSessionBaseURL, testUUIDString)

	expectedCode := http.StatusOK
	expectedType := ChallengeResponseSessionMediaType
	expectedBody := testCompleteSession

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		GetSession(testUUID, tenantID).
		Return([]byte(testCompleteSession), nil)

	v := mock_deps.NewMockIVerifier(ctrl)

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodGet, pathOK, http.NoBody)
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.Header.Set("Content-Type", testSupportedMediaTypeA)

	NewRouter(h).ServeHTTP(w, req)

	body := w.Body.Bytes()

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.JSONEq(t, expectedBody, string(body))
}

func TestHandler_DelSession_ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pathOK := path.Join(testSessionBaseURL, testUUIDString)

	expectedCode := http.StatusNoContent

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		DelSession(testUUID, tenantID).
		Return(nil)

	v := mock_deps.NewMockIVerifier(ctrl)

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodDelete, pathOK, http.NoBody)

	NewRouter(h).ServeHTTP(w, req)

	assert.Equal(t, expectedCode, w.Code)
}

func TestHandler_DelSession_bad_session_id(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	badPath := path.Join(testSessionBaseURL, "1234")

	expectedCode := http.StatusBadRequest
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
		Detail: fmt.Sprintf("invalid session id (%s) in path segment: invalid UUID length: 4", badPath),
	}

	sm := mock_deps.NewMockISessionManager(ctrl)
	v := mock_deps.NewMockIVerifier(ctrl)

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodDelete, badPath, http.NoBody)

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_DelSession_session_id_does_not_exist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pathOK := path.Join(testSessionBaseURL, testUUIDString)

	expectedCode := http.StatusInternalServerError
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Internal Server Error",
		Status: http.StatusInternalServerError,
		Detail: fmt.Sprintf("session id (%s) does not exist", testUUIDString),
	}

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		DelSession(testUUID, tenantID).
		Return(errors.New(`session id (` + testUUIDString + `) does not exist`))

	v := mock_deps.NewMockIVerifier(ctrl)

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodDelete, pathOK, http.NoBody)

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_GetWellKnownVerificationInfo_ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	supportedMediaTypes := []string{"application/type-1", "application/type-2"}

	sm := mock_deps.NewMockISessionManager(ctrl)

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		GetPublicKey().
		Return(&testKey, nil)
	v.EXPECT().
		SupportedMediaTypes().
		Return(supportedMediaTypes, nil)
	v.EXPECT().
		GetVTSState().
		Return(&testGoodServiceState, nil)

	expectedCode := http.StatusOK
	expectedType := capability.WellKnownMediaType
	expectedBody := capability.WellKnownInfo{
		MediaTypes:   supportedMediaTypes,
		Version:      testGoodServiceState.ServerVersion,
		ServiceState: capability.ServiceStateToAPI(testGoodServiceState.Status.String()),
		ApiEndpoints: publicApiMap,
	}

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodGet, "/.well-known/veraison/verification", http.NoBody)
	req.Header.Add("Accept", expectedType)

	NewRouter(h).ServeHTTP(w, req)

	var body capability.WellKnownInfo
	bytes := w.Body.Bytes()
	_ = json.Unmarshal(bytes, &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_GetWellKnownVerificationInfo_GetPublicKey_failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sm := mock_deps.NewMockISessionManager(ctrl)

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		GetPublicKey().
		Return(nil, errors.New("blah"))

	expectedCode := http.StatusInternalServerError
	expectedType := "application/problem+json"
	expectedErrorTitle := "Internal Server Error"

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodGet, "/.well-known/veraison/verification", http.NoBody)

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedErrorTitle, body.Title)
}

func TestHandler_GetWellKnownVerificationInfo_Get_SupportedMediaTypes_fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sm := mock_deps.NewMockISessionManager(ctrl)

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		GetPublicKey().
		Return(&testKey, nil)
	v.EXPECT().
		SupportedMediaTypes().
		Return(nil, errors.New("blah"))

	expectedCode := http.StatusInternalServerError
	expectedType := "application/problem+json"
	expectedErrorTitle := "Internal Server Error"

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodGet, "/.well-known/veraison/verification", http.NoBody)

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedErrorTitle, body.Title)
}

func TestHandler_GetWellKnownVerificationInfo_GetVTSState_fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sm := mock_deps.NewMockISessionManager(ctrl)
	supportedMediaTypes := []string{"application/type-1", "application/type-2"}

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		GetPublicKey().
		Return(&testKey, nil)
	v.EXPECT().
		SupportedMediaTypes().
		Return(supportedMediaTypes, nil)
	v.EXPECT().
		GetVTSState().
		Return(nil, errors.New("blah"))

	expectedCode := http.StatusInternalServerError
	expectedType := "application/problem+json"
	expectedErrorTitle := "Internal Server Error"

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodGet, "/.well-known/veraison/verification", http.NoBody)

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedErrorTitle, body.Title)
}

func TestHandler_GetWellKnownVerificationInfo_UnsupportedAccept(t *testing.T) {
	h := &Handler{}

	expectedCode := http.StatusNotAcceptable
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Not Acceptable",
		Status: http.StatusNotAcceptable,
		Detail: fmt.Sprintf("the only supported output format is %s", capability.WellKnownMediaType),
	}

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodGet, "/.well-known/veraison/verification", http.NoBody)
	g.Request.Header.Add("Accept", "application/unsupported+ber")

	NewRouter(h).ServeHTTP(w, g.Request)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func goodCMW(t *testing.T) []byte {
	w, err := cmw.NewMonad(testSupportedMediaTypeA, []byte(testJSONBody))
	require.NoError(t, err)
	b, err := w.MarshalJSON()
	require.NoError(t, err)
	return b
}

func TestHandler_SubmitEvidence_good_CMW(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pathOK := path.Join(testSessionBaseURL, testUUIDString)

	expectedCode := http.StatusOK
	expectedType := ChallengeResponseSessionMediaType

	testCMW := goodCMW(t)

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		GetSession(testUUID, tenantID).
		Return([]byte(testSession), nil)
	sm.EXPECT().
		SetSession(testUUID, tenantID, gomock.Any(), ConfigSessionTTL).
		Return(nil)

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		IsSupportedMediaType(testSupportedMediaTypeA).
		Return(true, nil)
	v.EXPECT().
		ProcessEvidence(tenantID, testNonce, []byte(testJSONBody), testSupportedMediaTypeA).
		Return([]byte(testResult), nil)

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, pathOK, bytes.NewReader(testCMW))
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.Header.Set("Content-Type", "application/vnd.veraison.cmw")

	NewRouter(h).ServeHTTP(w, req)

	_ = w.Body.Bytes()

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
}

func TestHandler_SubmitEvidence_bad_CMW(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	badCMW := []byte(`["missing value"]`)

	verifierError := "could not unwrap the CMW: wrong number of entries (1) in the CMW record"

	url := path.Join(testSessionBaseURL, testUUIDString)

	expectedCode := http.StatusBadRequest
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
		Detail: verifierError,
	}

	sm := mock_deps.NewMockISessionManager(ctrl)

	v := mock_deps.NewMockIVerifier(ctrl)

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(badCMW))
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.Header.Set("Content-Type", "application/vnd.veraison.cmw")

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}
