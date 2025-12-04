// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCcaPlatformWrapper_MarshalJSON(t *testing.T) {
	// Test with nil claims - this will panic as the underlying library expects non-nil claims
	// We test this to document the expected behavior
	wrapper := CcaPlatformWrapper{C: nil}
	
	assert.Panics(t, func() {
		_, _ = wrapper.MarshalJSON()
	}, "MarshalJSON should panic with nil claims")
}

func TestCcaRealmWrapper_MarshalJSON(t *testing.T) {
	// Test with nil claims - this will panic as the underlying library expects non-nil claims
	// We test this to document the expected behavior
	wrapper := CcaRealmWrapper{C: nil}
	
	assert.Panics(t, func() {
		_, _ = wrapper.MarshalJSON()
	}, "MarshalJSON should panic with nil claims")
}

func TestPsaPlatformWrapper_MarshalJSON(t *testing.T) {
	// Test with nil claims - this will panic as the underlying library expects non-nil claims
	// We test this to document the expected behavior
	wrapper := PsaPlatformWrapper{C: nil}
	
	assert.Panics(t, func() {
		_, _ = wrapper.MarshalJSON()
	}, "MarshalJSON should panic with nil claims")
}

func TestClaimsToMap(t *testing.T) {
	tests := []struct {
		name    string
		mapper  ClaimMapper
		wantErr bool
	}{
		{
			name: "nil CCA platform claims",
			mapper: CcaPlatformWrapper{C: nil},
			wantErr: true,
		},
		{
			name: "nil PSA platform claims", 
			mapper: PsaPlatformWrapper{C: nil},
			wantErr: true,
		},
		{
			name: "nil CCA realm claims",
			mapper: CcaRealmWrapper{C: nil},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				// These should panic due to nil claims
				assert.Panics(t, func() {
					_, _ = ClaimsToMap(tt.mapper)
				})
			} else {
				result, err := ClaimsToMap(tt.mapper)
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.IsType(t, map[string]interface{}{}, result)
			}
		})
	}
}

func TestMapToPSAClaims(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]interface{}
		wantErr bool
	}{
		{
			name: "invalid PSA claims map",
			input: map[string]interface{}{
				"invalid-field": "invalid-value",
			},
			wantErr: true,
		},
		{
			name:    "empty map",
			input:   map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MapToPSAClaims(tt.input)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestMapToCCAPlatformClaims(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]interface{}
		wantErr bool
	}{
		{
			name: "invalid CCA platform claims map",
			input: map[string]interface{}{
				"invalid-field": "invalid-value",
			},
			wantErr: true,
		},
		{
			name:    "empty map",
			input:   map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MapToCCAPlatformClaims(tt.input)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestGetImplID(t *testing.T) {
	tests := []struct {
		name     string
		scheme   string
		attr     json.RawMessage
		expected string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid implementation ID",
			scheme:   "test-scheme",
			attr:     json.RawMessage(`{"impl-id": "test-implementation-id"}`),
			expected: "test-implementation-id",
			wantErr:  false,
		},
		{
			name:    "missing impl-id field",
			scheme:  "test-scheme",
			attr:    json.RawMessage(`{"other-field": "value"}`),
			wantErr: true,
			errMsg:  "unable to get Implementation ID for scheme",
		},
		{
			name:    "impl-id is not a string",
			scheme:  "test-scheme",
			attr:    json.RawMessage(`{"impl-id": 123}`),
			wantErr: true,
			errMsg:  "unable to get Implementation ID for scheme",
		},
		{
			name:    "invalid JSON",
			scheme:  "test-scheme",
			attr:    json.RawMessage(`{invalid json}`),
			wantErr: true,
			errMsg:  "unable to get Implementation ID for scheme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetImplID(tt.scheme, tt.attr)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetInstID(t *testing.T) {
	tests := []struct {
		name     string
		scheme   string
		attr     json.RawMessage
		expected string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid instance ID",
			scheme:   "test-scheme",
			attr:     json.RawMessage(`{"inst-id": "test-instance-id"}`),
			expected: "test-instance-id",
			wantErr:  false,
		},
		{
			name:    "missing inst-id field",
			scheme:  "test-scheme", 
			attr:    json.RawMessage(`{"other-field": "value"}`),
			wantErr: true,
			errMsg:  "unable to get Instance ID",
		},
		{
			name:    "inst-id is not a string",
			scheme:  "test-scheme",
			attr:    json.RawMessage(`{"inst-id": 456}`),
			wantErr: true,
			errMsg:  "unable to get Instance ID",
		},
		{
			name:    "invalid JSON",
			scheme:  "test-scheme",
			attr:    json.RawMessage(`{invalid json}`),
			wantErr: true,
			errMsg:  "unable to get Instance ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetInstID(tt.scheme, tt.attr)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestDecodePemSubjectPubKeyInfo(t *testing.T) {
	// Generate a test RSA key pair
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	
	pubKey := &privKey.PublicKey
	
	// Convert to PKIX format and PEM encode
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	require.NoError(t, err)
	
	validPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	tests := []struct {
		name    string
		input   []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid PEM public key",
			input:   validPEM,
			wantErr: false,
		},
		{
			name:    "invalid PEM - no block",
			input:   []byte("not a pem block"),
			wantErr: true,
			errMsg:  "could not extract trust anchor PEM block",
		},
		{
			name:    "PEM with trailing data",
			input:   append(validPEM, []byte("trailing data")...),
			wantErr: true,
			errMsg:  "trailing data found after PEM block",
		},
		{
			name: "wrong PEM block type",
			input: pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: pubKeyBytes,
			}),
			wantErr: true,
			errMsg:  "unsupported key type",
		},
		{
			name: "invalid public key data",
			input: pem.EncodeToMemory(&pem.Block{
				Type:  "PUBLIC KEY",
				Bytes: []byte("invalid key data"),
			}),
			wantErr: true,
			errMsg:  "unable to parse public key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodePemSubjectPubKeyInfo(tt.input)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				
				// Verify the key is correct type
				parsedKey, ok := result.(*rsa.PublicKey)
				assert.True(t, ok)
				assert.Equal(t, pubKey.N, parsedKey.N)
				assert.Equal(t, pubKey.E, parsedKey.E)
			}
		})
	}
}