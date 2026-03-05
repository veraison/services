// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/go-cose"
)

func createTestCoservWithExpiry(t *testing.T, expiry time.Time, signed bool) []byte {
	resultSet := coserv.NewResultSet().SetExpiry(expiry)
	require.NotNil(t, resultSet)
	envSelector := coserv.NewEnvironmentSelector().AddInstance(
		coserv.StatefulInstance{
			Instance: comid.MustNewBytesInstance([]byte{0xde, 0xad, 0xbe, 0xef}),
		},
	)
	require.NotNil(t, envSelector)
	query, err := coserv.NewQuery(
		coserv.ArtifactTypeReferenceValues,
		*envSelector,
		coserv.ResultTypeBoth,
	)
	require.NoError(t, err)
	tv, err := coserv.NewCoserv("https://profile.example", *query)
	require.NoError(t, err)

	err = tv.AddResults(*resultSet)
	require.NoError(t, err)

	if signed {
		// For testing purposes, we can use a randomly generated key to sign the message
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err)
		signer, err := cose.NewSigner(cose.AlgorithmES256, privateKey)
		require.NoError(t, err)

		resultBytes, err := tv.Sign(signer)
		require.NoError(t, err)

		return resultBytes
	}

	resultBytes, err := cbor.Marshal(tv)
	require.NoError(t, err)

	return resultBytes
}

func testSetCacheHeaders_WithValidExpiry(t *testing.T, signed bool) {
	// Create a test result set with an expiry time 1 hour from now
	expiry := time.Now().Add(1 * time.Hour)
	tv := createTestCoservWithExpiry(t, expiry, signed)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/coserv/dGVzdAo", nil)

	setCacheHeaders(c, tv)

	cacheControl := w.Header().Get("Cache-Control")
	assert.NotEmpty(t, cacheControl)
	assert.Contains(t, cacheControl, "max-age=")

	var maxAge int64
	_, err := fmt.Sscanf(cacheControl, "max-age=%d", &maxAge)
	assert.NoError(t, err)
	// Allow some flexibility (within 10 seconds) to account for the time taken to execute the test
	assert.True(t, maxAge >= 3590 && maxAge <= 3610, "max-age should be approximately 3600 seconds")

	expiresHeader := w.Header().Get("Expires")
	assert.NotEmpty(t, expiresHeader)
	_, err = time.Parse(time.RFC1123, expiresHeader)
	assert.NoError(t, err, "Expires header should be in RFC1123 format")
}

func TestSetCacheHeaders_WithValidExpiryUnsigned(t *testing.T) {
	testSetCacheHeaders_WithValidExpiry(t, false)
}

func TestSetCacheHeaders_WithValidExpirySigned(t *testing.T) {
	testSetCacheHeaders_WithValidExpiry(t, true)
}

func testSetCacheHeaders_WithExpiredExpiry(t *testing.T, signed bool) {
	// Create a test result set with an expiry time in the past
	expiry := time.Now().Add(-1 * time.Hour)
	tv := createTestCoservWithExpiry(t, expiry, signed)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/coserv/dGVzdAo", nil)

	setCacheHeaders(c, tv)

	cacheControl := w.Header().Get("Cache-Control")
	assert.NotEmpty(t, cacheControl)
	assert.Contains(t, cacheControl, "max-age=0")
}

func TestSetCacheHeaders_WithExpiredExpiryUnsigned(t *testing.T) {
	testSetCacheHeaders_WithExpiredExpiry(t, false)
}

func TestSetCacheHeaders_WithExpiredExpirySigned(t *testing.T) {
	testSetCacheHeaders_WithExpiredExpiry(t, true)
}

func TestSetCacheHeaders_WithInvalidCBOR(t *testing.T) {
	// Pass invalid CBOR data
	invalidBytes := []byte{0xFF, 0xFF, 0xFF}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/coserv/dGVzdAo", nil)

	setCacheHeaders(c, invalidBytes)

	cacheControl := w.Header().Get("Cache-Control")
	assert.Empty(t, cacheControl)

	expiresHeader := w.Header().Get("Expires")
	assert.Empty(t, expiresHeader)
}
