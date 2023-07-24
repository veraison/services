// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/moogar0880/problems"
	"github.com/stretchr/testify/assert"
	"github.com/veraison/services/capability"
	"github.com/veraison/services/log"
	"github.com/veraison/services/proto"
	mock_deps "github.com/veraison/services/provisioning/api/mocks"
)

var (
	testGoodServiceState = proto.ServiceState{
		Status:        2,
		ServerVersion: "3.2",
	}
)

func TestHandler_Submit_UnsupportedAccept(t *testing.T) {
	h := &Handler{}

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

	dm := mock_deps.NewMockIProvisioner(ctrl)
	dm.EXPECT().
		IsSupportedMediaType(
			gomock.Eq(mediaType),
		).
		Return(false, nil)
	dm.EXPECT().
		SupportedMediaTypes().
		Return(supportedMediaTypes, nil)

	h := NewHandler(dm, log.Named("test"))

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

	dm := mock_deps.NewMockIProvisioner(ctrl)
	dm.EXPECT().
		IsSupportedMediaType(
			gomock.Eq(mediaType),
		).
		Return(true, nil)

	h := NewHandler(dm, log.Named("test"))

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

	dm := mock_deps.NewMockIProvisioner(ctrl)
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

	h := NewHandler(dm, log.Named("test"))

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
	dm := mock_deps.NewMockIProvisioner(ctrl)
	h := NewHandler(dm, log.Named("api"))

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

	dm := mock_deps.NewMockIProvisioner(ctrl)
	dm.EXPECT().
		SupportedMediaTypes().
		Return(supportedMediaTypes, nil)

	dm.EXPECT().
		GetVTSState().
		Return(&testGoodServiceState, nil)

	h := NewHandler(dm, log.Named("test"))

	expectedCode := http.StatusOK
	expectedType := capability.WellKnownMediaType
	expectedBody := capability.WellKnownInfo{
		MediaTypes:   supportedMediaTypes,
		Version:      testGoodServiceState.ServerVersion,
		ServiceState: capability.ServiceStateToAPI(testGoodServiceState.Status.String()),
		ApiEndpoints: publicApiMap,
	}

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodGet, "/.well-known/veraison/provisioning", http.NoBody)
	g.Request.Header.Add("Accept", expectedType)

	NewRouter(h).ServeHTTP(w, g.Request)

	var body capability.WellKnownInfo
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_GetWellKnownProvisioningInfo_GetRegisteredMediaTypes_empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dm := mock_deps.NewMockIProvisioner(ctrl)

	dm.EXPECT().
		SupportedMediaTypes().
		Return([]string{}, nil)

	dm.EXPECT().
		GetVTSState().
		Return(&testGoodServiceState, nil)

	h := NewHandler(dm, log.Named("test"))

	expectedCode := http.StatusOK
	expectedType := capability.WellKnownMediaType
	expectedBody := capability.WellKnownInfo{
		MediaTypes:   []string{},
		Version:      testGoodServiceState.ServerVersion,
		ServiceState: capability.ServiceStateToAPI(testGoodServiceState.Status.String()),
		ApiEndpoints: publicApiMap,
	}

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodGet, "/.well-known/veraison/provisioning", http.NoBody)
	g.Request.Header.Add("Accept", expectedType)

	NewRouter(h).ServeHTTP(w, g.Request)

	var body capability.WellKnownInfo
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_GetWellKnownProvisioningInfo_GetServiceState_fail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	supportedMediaTypes := []string{"application/type-1", "application/type-2"}
	dm := mock_deps.NewMockIProvisioner(ctrl)
	dm.EXPECT().
		SupportedMediaTypes().
		Return(supportedMediaTypes, nil)

	dm.EXPECT().
		GetVTSState().
		Return(nil, errors.New("blah"))

	h := NewHandler(dm, log.Named("test"))

	expectedCode := http.StatusInternalServerError
	expectedType := "application/problem+json"
	expectedErrorTitle := "Internal Server Error"

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodGet, "/.well-known/veraison/provisioning", http.NoBody)

	NewRouter(h).ServeHTTP(w, g.Request)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedErrorTitle, body.Title)
}

func TestHandler_GetWellKnownProvisioningInfo_UnsupportedAccept(t *testing.T) {
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

	g.Request, _ = http.NewRequest(http.MethodGet, "/.well-known/veraison/provisioning", http.NoBody)
	g.Request.Header.Add("Accept", "application/unsupported+ber")

	NewRouter(h).ServeHTTP(w, g.Request)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}
