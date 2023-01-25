// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package api

import (
	"bytes"
	"context"
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
	"github.com/veraison/services/decoder"
	"github.com/veraison/services/log"
	"github.com/veraison/services/proto"
	mock_deps "github.com/veraison/services/provisioning/api/mocks"
)

var (
	testGoodDecoderResponse = decoder.EndorsementDecoderResponse{
		TrustAnchors: []*proto.Endorsement{
			{},
		},
		ReferenceValues: []*proto.Endorsement{
			{},
		},
	}
	testFailedTaRes = proto.AddTrustAnchorResponse{
		Status: &proto.Status{Result: false},
	}
	testGoodTaRes = proto.AddTrustAnchorResponse{
		Status: &proto.Status{Result: true},
	}
	testFailedRefValRes = proto.AddRefValuesResponse{
		Status: &proto.Status{Result: false},
	}
	testGoodRefValRes = proto.AddRefValuesResponse{
		Status: &proto.Status{Result: true},
	}
)

type MockDecoder struct {
	Response *decoder.EndorsementDecoderResponse
}

func (o MockDecoder) Init(decoder.EndorsementDecoderParams) error { return nil }
func (o MockDecoder) Close() error                                { return nil }
func (o MockDecoder) GetName() string                             { return "mock" }
func (o MockDecoder) GetAttestationScheme() string                { return "mock" }
func (o MockDecoder) GetSupportedMediaTypes() []string            { return nil }

func (o MockDecoder) Decode(data []byte) (*decoder.EndorsementDecoderResponse, error) {
	return o.Response, nil
}

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

	dm := mock_deps.NewMockIManager[decoder.IEndorsementDecoder](ctrl)
	dm.EXPECT().
		IsRegisteredMediaType(
			gomock.Eq(mediaType),
		).
		Return(false)
	dm.EXPECT().
		GetRegisteredMediaTypes().
		Return(supportedMediaTypes)

	sc := mock_deps.NewMockIVTSClient(ctrl)

	h := NewHandler(dm, sc, log.Named("test"))

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

	dm := mock_deps.NewMockIManager[decoder.IEndorsementDecoder](ctrl)
	dm.EXPECT().
		IsRegisteredMediaType(
			gomock.Eq(mediaType),
		).
		Return(true)

	sc := mock_deps.NewMockIVTSClient(ctrl)

	h := NewHandler(dm, sc, log.Named("test"))

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
	decoderError := "decoder manager says: doh!"

	dm := mock_deps.NewMockIManager[decoder.IEndorsementDecoder](ctrl)
	dm.EXPECT().
		IsRegisteredMediaType(
			gomock.Eq(mediaType),
		).
		Return(true)
	dm.EXPECT().
		LookupByMediaType(
			gomock.Eq(mediaType),
		).
		Return(nil, errors.New(decoderError))

	sc := mock_deps.NewMockIVTSClient(ctrl)

	h := NewHandler(dm, sc, log.Named("test"))

	expectedCode := http.StatusOK
	expectedType := ProvisioningSessionMediaType
	expectedFailureReason := fmt.Sprintf("decoder manager returned error: %s", decoderError)
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

func TestHandler_Submit_store_AddTrustAnchor_failure1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mediaType := "application/good+json"
	endo := []byte("some data")
	storeError := "store says doh!"

	dm := mock_deps.NewMockIManager[decoder.IEndorsementDecoder](ctrl)
	dm.EXPECT().
		IsRegisteredMediaType(
			gomock.Eq(mediaType),
		).
		Return(true)
	dm.EXPECT().
		LookupByMediaType(
			gomock.Eq(mediaType),
		).
		Return(MockDecoder{&testGoodDecoderResponse}, nil)

	sc := mock_deps.NewMockIVTSClient(ctrl)
	sc.EXPECT().
		AddTrustAnchor(
			gomock.Eq(context.TODO()),
			gomock.Eq(
				&proto.AddTrustAnchorRequest{
					TrustAnchor: testGoodDecoderResponse.TrustAnchors[0],
				},
			),
		).
		Return(nil, errors.New(storeError))

	h := NewHandler(dm, sc, log.Named("test"))

	expectedCode := http.StatusOK
	expectedType := ProvisioningSessionMediaType
	expectedFailureReason := fmt.Sprintf(
		"endorsement store returned error: store operation failed for trust anchor: %s",
		storeError,
	)
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

func TestHandler_Submit_store_AddTrustAnchor_failure2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mediaType := "application/good+json"
	endo := []byte("some data")
	storeError := "store says doh!"
	testFailedTaRes.Status.ErrorDetail = storeError

	dm := mock_deps.NewMockIManager[decoder.IEndorsementDecoder](ctrl)
	dm.EXPECT().
		IsRegisteredMediaType(
			gomock.Eq(mediaType),
		).
		Return(true)
	dm.EXPECT().
		LookupByMediaType(
			gomock.Eq(mediaType),
		).
		Return(MockDecoder{&testGoodDecoderResponse}, nil)

	sc := mock_deps.NewMockIVTSClient(ctrl)
	sc.EXPECT().
		AddTrustAnchor(
			gomock.Eq(context.TODO()),
			gomock.Eq(
				&proto.AddTrustAnchorRequest{
					TrustAnchor: testGoodDecoderResponse.TrustAnchors[0],
				},
			),
		).
		Return(&testFailedTaRes, nil)

	h := NewHandler(dm, sc, log.Named("test"))

	expectedCode := http.StatusOK
	expectedType := ProvisioningSessionMediaType
	expectedFailureReason := fmt.Sprintf(
		"endorsement store returned error: store operation failed for trust anchor: %s",
		storeError,
	)
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

func TestHandler_Submit_store_AddRefValues_failure1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mediaType := "application/good+json"
	endo := []byte("some data")
	storeError := "store says doh!"

	dm := mock_deps.NewMockIManager[decoder.IEndorsementDecoder](ctrl)
	dm.EXPECT().
		IsRegisteredMediaType(
			gomock.Eq(mediaType),
		).
		Return(true)
	dm.EXPECT().
		LookupByMediaType(
			gomock.Eq(mediaType),
		).
		Return(MockDecoder{&testGoodDecoderResponse}, nil)

	sc := mock_deps.NewMockIVTSClient(ctrl)
	sc.EXPECT().
		AddTrustAnchor(
			gomock.Eq(context.TODO()),
			gomock.Eq(
				&proto.AddTrustAnchorRequest{
					TrustAnchor: testGoodDecoderResponse.TrustAnchors[0],
				},
			),
		).
		Return(&testGoodTaRes, nil)
	sc.EXPECT().
		AddRefValues(
			gomock.Eq(context.TODO()),
			gomock.Eq(
				&proto.AddRefValuesRequest{
					ReferenceValues: []*proto.Endorsement{
						{},
					},
				},
			),
		).
		Return(nil, errors.New(storeError))

	h := NewHandler(dm, sc, log.Named("test"))

	expectedCode := http.StatusOK
	expectedType := ProvisioningSessionMediaType
	expectedFailureReason := fmt.Sprintf(
		"endorsement store returned error: store operation failed for reference values: %s",
		storeError,
	)
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

func TestHandler_Submit_store_AddRefValues_failure2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mediaType := "application/good+json"
	endo := []byte("some data")
	storeError := "store says doh!"
	testFailedRefValRes.Status.ErrorDetail = storeError

	dm := mock_deps.NewMockIManager[decoder.IEndorsementDecoder](ctrl)
	dm.EXPECT().
		IsRegisteredMediaType(
			gomock.Eq(mediaType),
		).
		Return(true)
	dm.EXPECT().
		LookupByMediaType(
			gomock.Eq(mediaType),
		).
		Return(MockDecoder{&testGoodDecoderResponse}, nil)

	sc := mock_deps.NewMockIVTSClient(ctrl)
	sc.EXPECT().
		AddTrustAnchor(
			gomock.Eq(context.TODO()),
			gomock.Eq(
				&proto.AddTrustAnchorRequest{
					TrustAnchor: testGoodDecoderResponse.TrustAnchors[0],
				},
			),
		).
		Return(&testGoodTaRes, nil)
	sc.EXPECT().
		AddRefValues(
			gomock.Eq(context.TODO()),
			gomock.Eq(
				&proto.AddRefValuesRequest{
					ReferenceValues: []*proto.Endorsement{
						{},
					},
				},
			),
		).
		Return(&testFailedRefValRes, nil)

	h := NewHandler(dm, sc, log.Named("test"))

	expectedCode := http.StatusOK
	expectedType := ProvisioningSessionMediaType
	expectedFailureReason := fmt.Sprintf(
		"endorsement store returned error: store operation failed for reference values: %s",
		storeError,
	)
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

	dm := mock_deps.NewMockIManager[decoder.IEndorsementDecoder](ctrl)
	dm.EXPECT().
		IsRegisteredMediaType(
			gomock.Eq(mediaType),
		).
		Return(true)
	dm.EXPECT().
		LookupByMediaType(
			gomock.Eq(mediaType),
		).
		Return(MockDecoder{&testGoodDecoderResponse}, nil)

	sc := mock_deps.NewMockIVTSClient(ctrl)
	sc.EXPECT().
		AddTrustAnchor(
			gomock.Eq(context.TODO()),
			gomock.Eq(
				&proto.AddTrustAnchorRequest{
					TrustAnchor: testGoodDecoderResponse.TrustAnchors[0],
				},
			),
		).
		Return(&testGoodTaRes, nil)
	sc.EXPECT().
		AddRefValues(
			gomock.Eq(context.TODO()),
			gomock.Eq(
				&proto.AddRefValuesRequest{
					ReferenceValues: []*proto.Endorsement{
						{},
					},
				},
			),
		).
		Return(&testGoodRefValRes, nil)

	h := NewHandler(dm, sc, log.Named("test"))

	expectedCode := http.StatusOK
	expectedType := ProvisioningSessionMediaType
	expectedStatus := "success"

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
	assert.Nil(t, body.FailureReason)
	assert.Equal(t, expectedStatus, body.Status)
}
