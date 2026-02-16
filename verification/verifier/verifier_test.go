// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package verifier

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/veraison/services/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MockVTSClient is a mock implementation of vtsclient.IVTSClient for verifier tests
type MockVTSClient struct {
	ctrl                                     *gomock.Controller
	getSupportedVerificationMediaTypesErr   error
	getSupportedVerificationMediaTypesRes   *proto.MediaTypeList
	getAttestationErr                        error
	getAttestationRes                        *proto.AppraisalContext
	getServiceStateErr                       error
	getServiceStateRes                       *proto.ServiceState
}

func NewMockVTSClient(ctrl *gomock.Controller) *MockVTSClient {
	return &MockVTSClient{ctrl: ctrl}
}

func (m *MockVTSClient) GetSupportedVerificationMediaTypes(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*proto.MediaTypeList, error) {
	return m.getSupportedVerificationMediaTypesRes, m.getSupportedVerificationMediaTypesErr
}

func (m *MockVTSClient) GetAttestation(ctx context.Context, req *proto.AttestationToken, opts ...grpc.CallOption) (*proto.AppraisalContext, error) {
	return m.getAttestationRes, m.getAttestationErr
}

func (m *MockVTSClient) GetServiceState(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*proto.ServiceState, error) {
	return m.getServiceStateRes, m.getServiceStateErr
}

// Methods we don't need for verifier tests
func (m *MockVTSClient) GetSupportedProvisioningMediaTypes(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*proto.MediaTypeList, error) {
	return nil, errors.New("not implemented")
}

func (m *MockVTSClient) SubmitEndorsements(ctx context.Context, req *proto.SubmitEndorsementsRequest, opts ...grpc.CallOption) (*proto.SubmitEndorsementsResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *MockVTSClient) GetEARSigningPublicKey(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*proto.PublicKey, error) {
	return nil, errors.New("not implemented")
}

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	v := viper.New()
	mockClient := NewMockVTSClient(ctrl)
	verifier := New(v, mockClient)

	assert.NotNil(t, verifier)
	
	// Cast to concrete type to access VTSClient field
	ver, ok := verifier.(*Verifier)
	assert.True(t, ok)
	assert.Equal(t, mockClient, ver.VTSClient)
}

func TestVerifier_GetVTSState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name           string
		mockSetup      func(*MockVTSClient)
		expectedResult *proto.ServiceState
		expectedError  string
	}{
		{
			name: "successful state retrieval",
			mockSetup: func(m *MockVTSClient) {
				m.getServiceStateRes = &proto.ServiceState{
					Status:        proto.ServiceStatus_SERVICE_STATUS_READY,
					ServerVersion: "2.0.0",
				}
				m.getServiceStateErr = nil
			},
			expectedResult: &proto.ServiceState{
				Status:        proto.ServiceStatus_SERVICE_STATUS_READY,
				ServerVersion: "2.0.0",
			},
		},
		{
			name: "VTS client error",
			mockSetup: func(m *MockVTSClient) {
				m.getServiceStateRes = nil
				m.getServiceStateErr = errors.New("VTS service unavailable")
			},
			expectedResult: nil,
			expectedError:  "VTS service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockVTSClient(ctrl)
			tt.mockSetup(mockClient)
			
			verifier := &Verifier{VTSClient: mockClient}
			
			result, err := verifier.GetVTSState()
			
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestVerifier_IsSupportedMediaType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name           string
		inputMediaType string
		mockSetup      func(*MockVTSClient)
		expectedResult bool
		expectedError  string
	}{
		{
			name:           "supported media type",
			inputMediaType: "application/eat-cwt; profile=\"http://arm.com/psa/2.0.0\"",
			mockSetup: func(m *MockVTSClient) {
				m.getSupportedVerificationMediaTypesRes = &proto.MediaTypeList{
					MediaTypes: []string{
						"application/eat-cwt; profile=\"http://arm.com/psa/2.0.0\"",
						"application/psa-attestation-token",
					},
				}
				m.getSupportedVerificationMediaTypesErr = nil
			},
			expectedResult: true,
		},
		{
			name:           "unsupported media type",
			inputMediaType: "application/unknown-format",
			mockSetup: func(m *MockVTSClient) {
				m.getSupportedVerificationMediaTypesRes = &proto.MediaTypeList{
					MediaTypes: []string{
						"application/eat-cwt; profile=\"http://arm.com/psa/2.0.0\"",
						"application/psa-attestation-token",
					},
				}
				m.getSupportedVerificationMediaTypesErr = nil
			},
			expectedResult: false,
		},
		{
			name:           "invalid media type",
			inputMediaType: "",
			mockSetup: func(m *MockVTSClient) {
				// Mock setup not needed as validation should fail first
			},
			expectedResult: false,
			expectedError:  "invalid input parameter",
		},
		{
			name:           "VTS client error",
			inputMediaType: "application/json",
			mockSetup: func(m *MockVTSClient) {
				m.getSupportedVerificationMediaTypesRes = nil
				m.getSupportedVerificationMediaTypesErr = errors.New("VTS connection failed")
			},
			expectedResult: false,
			expectedError:  "VTS connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockVTSClient(ctrl)
			tt.mockSetup(mockClient)
			
			verifier := &Verifier{VTSClient: mockClient}
			
			result, err := verifier.IsSupportedMediaType(tt.inputMediaType)
			
			assert.Equal(t, tt.expectedResult, result)
			
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVerifier_SupportedMediaTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name           string
		mockSetup      func(*MockVTSClient)
		expectedResult []string
		expectedError  string
	}{
		{
			name: "successful retrieval",
			mockSetup: func(m *MockVTSClient) {
				m.getSupportedVerificationMediaTypesRes = &proto.MediaTypeList{
					MediaTypes: []string{
						"application/eat-cwt; profile=\"http://arm.com/psa/2.0.0\"",
						"application/psa-attestation-token",
					},
				}
				m.getSupportedVerificationMediaTypesErr = nil
			},
			expectedResult: []string{
				"application/eat-cwt; profile=\"http://arm.com/psa/2.0.0\"",
				"application/psa-attestation-token",
			},
		},
		{
			name: "VTS client error",
			mockSetup: func(m *MockVTSClient) {
				m.getSupportedVerificationMediaTypesRes = nil
				m.getSupportedVerificationMediaTypesErr = errors.New("VTS service error")
			},
			expectedResult: nil,
			expectedError:  "VTS service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockVTSClient(ctrl)
			tt.mockSetup(mockClient)
			
			verifier := &Verifier{VTSClient: mockClient}
			
			result, err := verifier.SupportedMediaTypes()
			
			assert.Equal(t, tt.expectedResult, result)
			
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVerifier_ProcessEvidence(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		tenantID      string
		nonce         []byte
		data          []byte
		mediaType     string
		mockSetup     func(*MockVTSClient)
		expectedResult []byte
		expectedError string
	}{
		{
			name:      "successful evidence processing",
			tenantID:  "tenant-123",
			nonce:     []byte("test-nonce"),
			data:      []byte("attestation-token-data"),
			mediaType: "application/psa-attestation-token",
			mockSetup: func(m *MockVTSClient) {
				m.getAttestationRes = &proto.AppraisalContext{
					Evidence: &proto.EvidenceContext{},
					Result:   []byte("processed-result"),
				}
				m.getAttestationErr = nil
			},
			expectedResult: nil, // The function returns ([]byte, error) but current implementation doesn't return processed data
		},
		{
			name:      "VTS client error during processing",
			tenantID:  "tenant-123",
			nonce:     []byte("test-nonce"),
			data:      []byte("attestation-token-data"),
			mediaType: "application/psa-attestation-token",
			mockSetup: func(m *MockVTSClient) {
				m.getAttestationRes = nil
				m.getAttestationErr = errors.New("attestation processing failed")
			},
			expectedResult: nil,
			expectedError:  "attestation processing failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockVTSClient(ctrl)
			tt.mockSetup(mockClient)
			
			verifier := &Verifier{VTSClient: mockClient}
			
			result, err := verifier.ProcessEvidence(tt.tenantID, tt.nonce, tt.data, tt.mediaType)
			
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				// The current implementation returns the appraisal context but we test that it doesn't error
			}
		})
	}
}