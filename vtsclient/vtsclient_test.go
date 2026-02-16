// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package vtsclient

import (
	"context"
	"errors"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/veraison/services/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestNewGRPC(t *testing.T) {
	client := NewGRPC()
	
	assert.NotNil(t, client)
	assert.Empty(t, client.ServerAddress)
	assert.Nil(t, client.Credentials)
	assert.Nil(t, client.Connection)
}

func TestGRPC_Init(t *testing.T) {
	tests := []struct {
		name        string
		setupViper  func() *viper.Viper
		certPath    string
		keyPath     string
		expectError bool
		errorSubstr string
	}{
		{
			name: "successful init with insecure credentials",
			setupViper: func() *viper.Viper {
				v := viper.New()
				v.Set("server-addr", "localhost:50051")
				v.Set("tls", false)
				return v
			},
			certPath:    "",
			keyPath:     "",
			expectError: false,
		},
		{
			name: "init with TLS enabled but no cert files",
			setupViper: func() *viper.Viper {
				v := viper.New()
				v.Set("server-addr", "localhost:50051") 
				v.Set("tls", true)
				return v
			},
			certPath:    "/nonexistent/cert.pem",
			keyPath:     "/nonexistent/key.pem",
			expectError: true,
			errorSubstr: "no such file",
		},
		{
			name: "missing server address",
			setupViper: func() *viper.Viper {
				v := viper.New()
				// Don't set server-addr
				v.Set("tls", false)
				return v
			},
			certPath:    "",
			keyPath:     "",
			expectError: false, // Should not error, just have empty server address
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewGRPC()
			v := tt.setupViper()
			
			err := client.Init(v, tt.certPath, tt.keyPath)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorSubstr != "" {
					assert.Contains(t, err.Error(), tt.errorSubstr)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client.Credentials)
				
				// For insecure connections, just verify credentials exist
				if !v.GetBool("tls") {
					assert.NotNil(t, client.Credentials)
				}
			}
		})
	}
}

func TestNoConnectionError(t *testing.T) {
	originalErr := errors.New("connection failed")
	err := NewNoConnectionError("test-context", originalErr)
	
	assert.Equal(t, "test-context", err.Context)
	assert.Equal(t, originalErr, err.Err)
	
	expectedMsg := "(from: test-context) connection failed"
	assert.Equal(t, expectedMsg, err.Error())
	
	assert.Equal(t, originalErr, err.Unwrap())
}

func TestGRPC_EnsureConnection_NoAddress(t *testing.T) {
	client := NewGRPC()
	// Don't set ServerAddress or Credentials
	
	err := client.EnsureConnection()
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection to gRPC VTS server")
}

func TestGRPC_EnsureConnection_AlreadyConnected(t *testing.T) {
	// This test would require more complex mocking setup
	t.Skip("Skipping due to difficulty mocking grpc.ClientConn")
}

func TestGRPC_GetProvisionerClient_NoConnection(t *testing.T) {
	client := NewGRPC()
	
	provisionerClient := client.GetProvisionerClient()
	
	assert.Nil(t, provisionerClient)
}

func TestGRPC_GetProvisionerClient_WithConnection(t *testing.T) {
	// Skip this test due to mocking complexity
	t.Skip("Skipping due to difficulty mocking grpc.ClientConn")
}

func TestGRPC_ServiceMethods_NoConnection(t *testing.T) {
	client := NewGRPC()
	ctx := context.Background()
	
	tests := []struct {
		name   string
		method func() error
	}{
		{
			name: "GetServiceState",
			method: func() error {
				_, err := client.GetServiceState(ctx, &emptypb.Empty{})
				return err
			},
		},
		{
			name: "GetAttestation",
			method: func() error {
				_, err := client.GetAttestation(ctx, &proto.AttestationToken{})
				return err
			},
		},
		{
			name: "GetSupportedVerificationMediaTypes",
			method: func() error {
				_, err := client.GetSupportedVerificationMediaTypes(ctx, &emptypb.Empty{})
				return err
			},
		},
		{
			name: "GetSupportedProvisioningMediaTypes",
			method: func() error {
				_, err := client.GetSupportedProvisioningMediaTypes(ctx, &emptypb.Empty{})
				return err
			},
		},
		{
			name: "SubmitEndorsements",
			method: func() error {
				_, err := client.SubmitEndorsements(ctx, &proto.SubmitEndorsementsRequest{})
				return err
			},
		},
		{
			name: "GetEARSigningPublicKey",
			method: func() error {
				_, err := client.GetEARSigningPublicKey(ctx, &emptypb.Empty{})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.method()
			
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "connection to gRPC VTS server")
		})
	}
}

func TestGRPC_ServiceMethods_NoClient(t *testing.T) {
	// Skip this test due to complexity of mocking grpc.ClientConn
	t.Skip("Skipping due to difficulty mocking grpc.ClientConn for nil client scenario")
}

func TestNormalizeMediaTypeList(t *testing.T) {
	tests := []struct {
		name     string
		input    *proto.MediaTypeList
		expected []string
	}{
		{
			name: "valid media types",
			input: &proto.MediaTypeList{
				MediaTypes: []string{
					"application/json",
					"application/cbor",
					"text/plain",
				},
			},
			expected: []string{
				"application/json",
				"application/cbor", 
				"text/plain",
			},
		},
		{
			name: "empty media type list",
			input: &proto.MediaTypeList{
				MediaTypes: []string{},
			},
			expected: []string{},
		},
		{
			name: "mixed valid and invalid media types",
			input: &proto.MediaTypeList{
				MediaTypes: []string{
					"application/json",
					"invalid/media/type/with/too/many/slashes",
					"application/cbor",
				},
			},
			expected: []string{
				"application/json",
				"application/cbor",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeMediaTypeList(tt.input)
			
			assert.NotNil(t, result)
			assert.Equal(t, len(tt.expected), len(result.MediaTypes))
			
			for i, expected := range tt.expected {
				assert.Equal(t, expected, result.MediaTypes[i])
			}
		})
	}
}

// Note: More complex mocking of grpc.ClientConn would require additional testing infrastructure
// The key functionality we can test is initialization, error handling, and connection setup logic