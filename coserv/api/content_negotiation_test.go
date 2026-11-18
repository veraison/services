// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/services/proto"
)

const (
	testProfile      = "tag:arm.com,2025:cca_platform#1.0.0"
	testBadProfile   = "tag:invalid.com,2025:cca_platform#1.0.0"
	testCborMT       = `application/coserv+cbor; profile="` + testProfile + `"`
	testCoseMT       = `application/coserv+cose; profile="` + testProfile + `"`
	testSupportedCbr = testCborMT
	testSupportedCse = testCoseMT
)

var testSupported = []string{testSupportedCbr, testSupportedCse}

func TestSplitAccept_CommaInsideQuotes(t *testing.T) {
	parts := splitAccept(`application/coserv+cbor; profile="tag:arm.com,2025:cca_platform#1.0.0", application/json`)
	require.Len(t, parts, 2)
	assert.Contains(t, parts[0], "tag:arm.com,2025")
}

func TestNegotiateCoserv(t *testing.T) {
	tests := []struct {
		name    string
		accept  string
		profile string
		want    string // empty means an error is expected
	}{
		{"no accept header uses first supported", "", testProfile, testCborMT},
		{"plain cbor", "application/coserv+cbor", testProfile, testCborMT},
		{"plain cose", "application/coserv+cose", testProfile, testCoseMT},
		{"wildcard", "*/*", testProfile, testCborMT},
		{"matching accept profile", testCoseMT, testProfile, testCoseMT},
		{"q-values prefer cose",
			"application/coserv+cbor; q=0.5, application/coserv+cose; q=1.0",
			testProfile, testCoseMT},
		{"q=0 is not acceptable", "application/coserv+cbor; q=0", testProfile, ""},
		{"unsupported accept profile",
			`application/coserv+cbor; profile="` + testBadProfile + `"`, testProfile, ""},
		{"unsupported query profile", "application/coserv+cbor", testBadProfile, ""},
		{"accept names no coserv type", "application/json", testProfile, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := negotiateMediaType(tt.accept, testSupported, tt.profile)
			if tt.want == "" {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNegotiateCoserv_CoseNotOffered(t *testing.T) {
	// no CoSERV signer configured: only +cbor is advertised
	_, err := negotiateMediaType("application/coserv+cose", []string{testCborMT}, testProfile)
	assert.Error(t, err)
}

// fakeDistributor lets the handler be exercised without VTS.
type fakeDistributor struct {
	supported []string
	calledMT  string
	result    []byte
}

func (f *fakeDistributor) SupportedMediaTypes() ([]string, error)  { return f.supported, nil }
func (f *fakeDistributor) GetPublicKey() (*proto.PublicKey, error) { return nil, nil }
func (f *fakeDistributor) GetEndorsements(_, _, mediaType string) ([]byte, error) {
	f.calledMT = mediaType
	return f.result, nil
}

func newQuery(t *testing.T, profile string) string {
	sel := coserv.NewEnvironmentSelector().AddInstance(
		coserv.StatefulInstance{Instance: comid.MustNewBytesInstance([]byte{0xde, 0xad, 0xbe, 0xef})},
	)
	q, err := coserv.NewEnvironmentQuery(coserv.ArtifactTypeReferenceValues, *sel, coserv.ResultTypeBoth)
	require.NoError(t, err)
	cs, err := coserv.NewCoserv(profile, *q)
	require.NoError(t, err)
	b64, err := cs.ToBase64Url()
	require.NoError(t, err)

	return b64
}

func serve(fd *fakeDistributor, query, accept string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)

	h := Handler{EndorsementDistributor: fd}
	r := gin.New()
	r.GET("/coserv/:query", h.CoservRequest)

	req := httptest.NewRequest(http.MethodGet, "/coserv/"+query, http.NoBody)
	if accept != "" {
		req.Header.Set("Accept", accept)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

func TestCoservRequest_Profiles(t *testing.T) {
	result := createTestCoservWithExpiry(t, time.Now().Add(time.Hour), false)

	t.Run("supported profile returns 200", func(t *testing.T) {
		fd := &fakeDistributor{supported: testSupported, result: result}
		w := serve(fd, newQuery(t, testProfile), "application/coserv+cbor")
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, testCborMT, fd.calledMT)
	})

	t.Run("unsupported accept profile returns 406", func(t *testing.T) {
		fd := &fakeDistributor{supported: testSupported, result: result}
		w := serve(fd, newQuery(t, testProfile),
			`application/coserv+cbor; profile="`+testBadProfile+`"`)
		assert.Equal(t, http.StatusNotAcceptable, w.Code)
		assert.Empty(t, fd.calledMT, "VTS must not be queried")
	})

	t.Run("unsupported query profile returns 406", func(t *testing.T) {
		fd := &fakeDistributor{supported: testSupported, result: result}
		w := serve(fd, newQuery(t, testBadProfile), "application/coserv+cbor")
		assert.Equal(t, http.StatusNotAcceptable, w.Code)
		assert.Empty(t, fd.calledMT, "VTS must not be queried")
	})

	t.Run("non-coserv accept returns 406", func(t *testing.T) {
		fd := &fakeDistributor{supported: testSupported, result: result}
		w := serve(fd, newQuery(t, testProfile), "application/json")
		assert.Equal(t, http.StatusNotAcceptable, w.Code)
	})

	t.Run("malformed query still returns 400", func(t *testing.T) {
		fd := &fakeDistributor{supported: testSupported, result: result}
		w := serve(fd, "not-a-query", "application/coserv+cbor")
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
