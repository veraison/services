// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package provisioning

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/moogar0880/problems"
	"github.com/stretchr/testify/assert"
	"github.com/veraison/services/pkg/api/common"
	"github.com/veraison/services/pkg/api/common/auth"
	"github.com/veraison/services/pkg/log"
	"github.com/veraison/services/pkg/vts"
)

var (
	testGoodServiceState = vts.State{
		Version: "3.2",
	}
	maxPayloadSize = int64(99999)
)

func TestHandler_Submit_UnsupportedAccept(t *testing.T) {
	h := &Handler{logger: log.Named("test")}

	expectedCode := http.StatusNotAcceptable
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Not Acceptable",
		Status: http.StatusNotAcceptable,
		Detail: fmt.Sprintf("the only supported output format is %s", ProvisioningSessionMediaType),
	}

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodPost, "/", http.NoBody)
	g.Request.Header.Add("Accept", "application/unsupported+ber")

	h.Submit(g)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_Submit_UnsupportedMediaType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mediaType := "application/unsupported+json"
	supportedMediaTypes := []string{"application/type-1", "application/type-2"}

	dm := NewMockIProvisioner(ctrl)
	dm.EXPECT().
		IsSupportedMediaType(
			gomock.Eq(mediaType),
		).
		Return(false, nil)
	dm.EXPECT().
		SupportedMediaTypes().
		Return(supportedMediaTypes, nil)

	h := NewHandler(dm, log.Named("test"), time.Hour, maxPayloadSize)

	expectedCode := http.StatusUnsupportedMediaType
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Unsupported Media Type",
		Status: http.StatusUnsupportedMediaType,
		Detail: fmt.Sprintf("no active plugin found for %s", mediaType),
	}
	expectedAcceptHeader := strings.Join(supportedMediaTypes, ", ")

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodPost, "/", http.NoBody)
	g.Request.Header.Add("Content-Type", mediaType)
	g.Request.Header.Add("Accept", ProvisioningSessionMediaType)

	h.Submit(g)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedAcceptHeader, w.Result().Header.Get("Accept"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_Submit_NoBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mediaType := "application/good+json"

	dm := NewMockIProvisioner(ctrl)
	dm.EXPECT().
		IsSupportedMediaType(
			gomock.Eq(mediaType),
		).
		Return(true, nil)

	h := NewHandler(dm, log.Named("test"), time.Hour, maxPayloadSize)

	expectedCode := http.StatusBadRequest
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
		Detail: "empty body",
	}

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	emptyBody := []byte("")

	g.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(emptyBody))
	g.Request.Header.Add("Content-Type", mediaType)
	g.Request.Header.Add("Accept", ProvisioningSessionMediaType)

	h.Submit(g)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_Submit_DecodeFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mediaType := "application/good+json"
	endo := []byte("some data")
	handlerError := "decoding failure: doh!"

	dm := NewMockIProvisioner(ctrl)
	dm.EXPECT().
		IsSupportedMediaType(
			gomock.Eq(mediaType),
		).
		Return(true, nil)
	dm.EXPECT().
		SubmitEndorsements(
			tenantID, endo, gomock.Eq(mediaType),
		).
		Return(errors.New(handlerError))

	h := NewHandler(dm, log.Named("test"), time.Hour, maxPayloadSize)

	expectedCode := http.StatusOK
	expectedType := ProvisioningSessionMediaType
	expectedFailureReason := fmt.Sprintf("submit endorsement returned error: %s", handlerError)
	expectedStatus := "failed"

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(endo))
	g.Request.Header.Add("Content-Type", mediaType)
	g.Request.Header.Add("Accept", ProvisioningSessionMediaType)

	h.Submit(g)

	var body ProvisioningSession
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.NotNil(t, body.FailureReason)
	assert.Equal(t, expectedFailureReason, *body.FailureReason)
	assert.Equal(t, expectedStatus, body.Status)
}

func TestHandler_Submit_ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mediaType := "application/good+json"
	endo := []byte("some data")
	expectedCode := http.StatusOK
	expectedType := ProvisioningSessionMediaType
	expectedStatus := "success"
	dm := NewMockIProvisioner(ctrl)
	h := NewHandler(dm, log.Named("api"), time.Hour, maxPayloadSize)

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	dm.EXPECT().
		IsSupportedMediaType(
			gomock.Eq(mediaType),
		).
		Return(true, nil)
	dm.EXPECT().
		SubmitEndorsements(
			tenantID, endo, gomock.Eq(mediaType),
		).
		Return(nil)
	g.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(endo))
	g.Request.Header.Add("Content-Type", mediaType)
	g.Request.Header.Add("Accept", ProvisioningSessionMediaType)

	h.Submit(g)

	var body ProvisioningSession
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Nil(t, body.FailureReason)
	assert.Equal(t, expectedStatus, body.Status)
}

func TestHandler_GetWellKnownProvisioningInfo_ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	supportedMediaTypes := []string{"application/type-1", "application/type-2"}

	dm := NewMockIProvisioner(ctrl)
	dm.EXPECT().
		SupportedMediaTypes().
		Return(supportedMediaTypes, nil)

	dm.EXPECT().
		ServerVersion().
		Return(testGoodServiceState.Version, nil)

	h := NewHandler(dm, log.Named("test"), time.Hour, maxPayloadSize)

	expectedCode := http.StatusOK
	expectedType := common.WellKnownMediaType
	expectedBody := common.WellKnownInfo{
		MediaTypes:   supportedMediaTypes,
		Version:      testGoodServiceState.Version,
		ApiEndpoints: publicApiMap,
	}

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodGet, "/.well-known/veraison/provisioning", http.NoBody)
	g.Request.Header.Add("Accept", expectedType)

	u := auth.NewPassthroughAuthorizer(log.Named("auth"))
	SetupRouter(gin.Default(), h, u).ServeHTTP(w, g.Request)

	var body common.WellKnownInfo
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_GetWellKnownProvisioningInfo_GetRegisteredMediaTypes_empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dm := NewMockIProvisioner(ctrl)

	dm.EXPECT().
		SupportedMediaTypes().
		Return([]string{}, nil)

	dm.EXPECT().
		ServerVersion().
		Return(testGoodServiceState.Version, nil)

	h := NewHandler(dm, log.Named("test"), time.Hour, maxPayloadSize)

	expectedCode := http.StatusOK
	expectedType := common.WellKnownMediaType
	expectedBody := common.WellKnownInfo{
		Version:      testGoodServiceState.Version,
		ApiEndpoints: publicApiMap,
	}

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodGet, "/.well-known/veraison/provisioning", http.NoBody)
	g.Request.Header.Add("Accept", expectedType)

	u := auth.NewPassthroughAuthorizer(log.Named("auth"))
	SetupRouter(gin.Default(), h, u).ServeHTTP(w, g.Request)

	var body common.WellKnownInfo
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_GetWellKnownProvisioningInfo_GetServiceState_fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	supportedMediaTypes := []string{"application/type-1", "application/type-2"}
	dm := NewMockIProvisioner(ctrl)
	dm.EXPECT().
		SupportedMediaTypes().
		Return(supportedMediaTypes, nil)

	dm.EXPECT().
		ServerVersion().
		Return("", errors.New("blah"))

	h := NewHandler(dm, log.Named("test"), time.Hour, maxPayloadSize)

	expectedCode := http.StatusInternalServerError
	expectedType := "application/problem+json"
	expectedErrorTitle := "Internal Server Error"

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodGet, "/.well-known/veraison/provisioning", http.NoBody)

	u := auth.NewPassthroughAuthorizer(log.Named("auth"))
	SetupRouter(gin.Default(), h, u).ServeHTTP(w, g.Request)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedErrorTitle, body.Title)
}

func TestHandler_GetWellKnownProvisioningInfo_UnsupportedAccept(t *testing.T) {
	h := &Handler{logger: log.Named("test")}

	expectedCode := http.StatusNotAcceptable
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Not Acceptable",
		Status: http.StatusNotAcceptable,
		Detail: fmt.Sprintf("the only supported output format is %s", common.WellKnownMediaType),
	}

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodGet, "/.well-known/veraison/provisioning", http.NoBody)
	g.Request.Header.Add("Accept", "application/unsupported+ber")

	u := auth.NewPassthroughAuthorizer(log.Named("auth"))
	SetupRouter(gin.Default(), h, u).ServeHTTP(w, g.Request)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_Submit_payload_too_large(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mediaType := "application/good+json"
	endo := []byte("some data")
	expectedCode := http.StatusBadRequest
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
		Detail: "error reading body: http: request body too large",
	}

	dm := NewMockIProvisioner(ctrl)
	h := NewHandler(dm, log.Named("api"), time.Hour, 1)

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	dm.EXPECT().
		IsSupportedMediaType(
			gomock.Eq(mediaType),
		).
		Return(true, nil)
	g.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(endo))
	g.Request.Header.Add("Content-Type", mediaType)
	g.Request.Header.Add("Accept", ProvisioningSessionMediaType)

	h.Submit(g)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_SetActive_UnsupportedMediaType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dm := NewMockIProvisioner(ctrl)
	h := NewHandler(dm, log.Named("test"), time.Hour, maxPayloadSize)

	mediaType := "application/unsupported+json"

	expectedCode := http.StatusUnsupportedMediaType
	expectedType := "application/concise-problem-details+cbor"
	expectedBody := common.ConciseProblem{
		Title:  fmt.Sprintf("%d %v", expectedCode, http.StatusText(expectedCode)),
		Detail: fmt.Sprintf("unknown media type: %v, expected: %v", mediaType, ELMQueryMediaType),
	}

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodPost, "/", http.NoBody)
	g.Request.Header.Add("Content-Type", mediaType)

	h.SetActive(true)(g)

	var body common.ConciseProblem
	_ = cbor.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_SetActive_NoBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dm := NewMockIProvisioner(ctrl)
	h := NewHandler(dm, log.Named("test"), time.Hour, maxPayloadSize)

	expectedCode := http.StatusBadRequest
	expectedType := "application/concise-problem-details+cbor"
	expectedBody := common.ConciseProblem{
		Title:  fmt.Sprintf("%d %v", expectedCode, http.StatusText(expectedCode)),
		Detail: "empty body",
	}

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte{}))
	g.Request.Header.Add("Content-Type", ELMQueryMediaType)

	h.SetActive(true)(g)

	var body common.ConciseProblem
	_ = cbor.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_SetActive_ProvisionerFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mediaType := ELMQueryMediaType
	elm := []byte("some query")
	handlerError := "provisioner failure: uh oh!"

	dm := NewMockIProvisioner(ctrl)
	dm.EXPECT().
		SetEndorsementsState(
			tenantID, elm, true,
		).
		Return(errors.New(handlerError))

	h := NewHandler(dm, log.Named("test"), time.Hour, maxPayloadSize)

	expectedCode := http.StatusBadRequest
	expectedType := "application/concise-problem-details+cbor"
	expectedBody := common.ConciseProblem{
		Title:  fmt.Sprintf("%d %v", expectedCode, http.StatusText(expectedCode)),
		Detail: handlerError,
	}

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(elm))
	g.Request.Header.Add("Content-Type", mediaType)

	h.SetActive(true)(g)

	var body common.ConciseProblem
	_ = cbor.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_SetActive_ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mediaType := ELMQueryMediaType
	elm := []byte("some query")

	dm := NewMockIProvisioner(ctrl)
	dm.EXPECT().
		SetEndorsementsState(
			tenantID, elm, true,
		).
		Return(nil)

	h := NewHandler(dm, log.Named("test"), time.Hour, maxPayloadSize)

	expectedCode := http.StatusNoContent

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(elm))
	g.Request.Header.Add("Content-Type", mediaType)

	h.SetActive(true)(g)

	assert.Equal(t, expectedCode, w.Code)
	assert.Nil(t, w.Body.Bytes())
}
