// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package provisioner

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/vtsclient"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MockVTSClient is a mock implementation of vtsclient.IVTSClient
type MockVTSClient struct {
	ctrl                                  *gomock.Controller
	getSupportedProvisioningMediaTypesErr error
	getSupportedProvisioningMediaTypesRes *proto.MediaTypeList
	submitEndorsementsErr                 error
	submitEndorsementsRes                 *proto.SubmitEndorsementsResponse
	getServiceStateErr                    error
	getServiceStateRes                    *proto.ServiceState
}

func NewMockVTSClient(ctrl *gomock.Controller) *MockVTSClient {
	return &MockVTSClient{ctrl: ctrl}
}

func (m *MockVTSClient) GetSupportedProvisioningMediaTypes(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*proto.MediaTypeList, error) {
	return m.getSupportedProvisioningMediaTypesRes, m.getSupportedProvisioningMediaTypesErr
}

func (m *MockVTSClient) SubmitEndorsements(ctx context.Context, req *proto.SubmitEndorsementsRequest, opts ...grpc.CallOption) (*proto.SubmitEndorsementsResponse, error) {
	return m.submitEndorsementsRes, m.submitEndorsementsErr
}

func (m *MockVTSClient) GetServiceState(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*proto.ServiceState, error) {
	return m.getServiceStateRes, m.getServiceStateErr
}

// Methods we don't need for provisioner tests
func (m *MockVTSClient) GetAttestation(ctx context.Context, req *proto.AttestationToken, opts ...grpc.CallOption) (*proto.AppraisalContext, error) {
	return nil, errors.New("not implemented")
}

func (m *MockVTSClient) GetSupportedVerificationMediaTypes(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*proto.MediaTypeList, error) {
	return nil, errors.New("not implemented")
}

func (m *MockVTSClient) GetEARSigningPublicKey(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*proto.PublicKey, error) {
	return nil, errors.New("not implemented")
}

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := NewMockVTSClient(ctrl)
	provisioner := New(mockClient)

	assert.NotNil(t, provisioner)
	
	// Cast to concrete type to access VTSClient field
	p, ok := provisioner.(*Provisioner)
	assert.True(t, ok)
	assert.Equal(t, mockClient, p.VTSClient)
}

func TestProvisioner_IsSupportedMediaType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		inputMediaType string
		mockSetup      func(*MockVTSClient)
		expectedResult bool
		expectedError  string
	}{
		{
			name:           "supported media type",
			inputMediaType: "application/json",
			mockSetup: func(m *MockVTSClient) {
				m.getSupportedProvisioningMediaTypesRes = &proto.MediaTypeList{
					MediaTypes: []string{"application/json", "application/cbor"},
				}
				m.getSupportedProvisioningMediaTypesErr = nil
			},
			expectedResult: true,
		},
		{
			name:           "unsupported media type",
			inputMediaType: "application/xml",
			mockSetup: func(m *MockVTSClient) {
				m.getSupportedProvisioningMediaTypesRes = &proto.MediaTypeList{
					MediaTypes: []string{"application/json", "application/cbor"},
				}
				m.getSupportedProvisioningMediaTypesErr = nil
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
				m.getSupportedProvisioningMediaTypesRes = nil
				m.getSupportedProvisioningMediaTypesErr = errors.New("VTS error")
			},
			expectedResult: false,
			expectedError:  "VTS error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockVTSClient(ctrl)
			tt.mockSetup(mockClient)
			
			provisioner := &Provisioner{VTSClient: mockClient}
			
			result, err := provisioner.IsSupportedMediaType(tt.inputMediaType)
			
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

func TestProvisioner_SupportedMediaTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name            string
		mockSetup       func(*MockVTSClient)
		expectedResult  []string
		expectedError   string
	}{
		{
			name: "successful retrieval",
			mockSetup: func(m *MockVTSClient) {
				m.getSupportedProvisioningMediaTypesRes = &proto.MediaTypeList{
					MediaTypes: []string{"application/json", "application/cbor"},
				}
				m.getSupportedProvisioningMediaTypesErr = nil
			},
			expectedResult: []string{"application/json", "application/cbor"},
		},
		{
			name: "VTS client error",
			mockSetup: func(m *MockVTSClient) {
				m.getSupportedProvisioningMediaTypesRes = nil
				m.getSupportedProvisioningMediaTypesErr = errors.New("VTS connection error")
			},
			expectedResult: nil,
			expectedError:  "VTS connection error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockVTSClient(ctrl)
			tt.mockSetup(mockClient)
			
			provisioner := &Provisioner{VTSClient: mockClient}
			
			result, err := provisioner.SupportedMediaTypes()
			
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

func TestProvisioner_SubmitEndorsements(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		tenantID      string
		data          []byte
		mediaType     string
		mockSetup     func(*MockVTSClient)
		expectedError string
	}{
		{
			name:      "successful submission",
			tenantID:  "tenant1",
			data:      []byte("test data"),
			mediaType: "application/json",
			mockSetup: func(m *MockVTSClient) {
				m.submitEndorsementsRes = &proto.SubmitEndorsementsResponse{
					Status: &proto.Status{Result: true},
				}
				m.submitEndorsementsErr = nil
			},
		},
		{
			name:      "VTS client connection error",
			tenantID:  "tenant1",
			data:      []byte("test data"),
			mediaType: "application/json",
			mockSetup: func(m *MockVTSClient) {
				m.submitEndorsementsRes = nil
				m.submitEndorsementsErr = vtsclient.NewNoConnectionError("test", errors.New("connection failed"))
			},
			expectedError: "no connection",
		},
		{
			name:      "VTS client other error",
			tenantID:  "tenant1", 
			data:      []byte("test data"),
			mediaType: "application/json",
			mockSetup: func(m *MockVTSClient) {
				m.submitEndorsementsRes = nil
				m.submitEndorsementsErr = errors.New("internal error")
			},
			expectedError: "submit endorsements failed: internal error",
		},
		{
			name:      "submission failed with error detail",
			tenantID:  "tenant1",
			data:      []byte("test data"),
			mediaType: "application/json",
			mockSetup: func(m *MockVTSClient) {
				m.submitEndorsementsRes = &proto.SubmitEndorsementsResponse{
					Status: &proto.Status{
						Result:      false,
						ErrorDetail: "validation failed",
					},
				}
				m.submitEndorsementsErr = nil
			},
			expectedError: "submit endorsements failed: validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockVTSClient(ctrl)
			tt.mockSetup(mockClient)
			
			provisioner := &Provisioner{VTSClient: mockClient}
			
			err := provisioner.SubmitEndorsements(tt.tenantID, tt.data, tt.mediaType)
			
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProvisioner_GetVTSState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name          string
		mockSetup     func(*MockVTSClient)
		expectedResult *proto.ServiceState
		expectedError string
	}{
		{
			name: "successful state retrieval",
			mockSetup: func(m *MockVTSClient) {
				m.getServiceStateRes = &proto.ServiceState{
					Status:        proto.ServiceStatus_SERVICE_STATUS_READY,
					ServerVersion: "1.0.0",
				}
				m.getServiceStateErr = nil
			},
			expectedResult: &proto.ServiceState{
				Status:        proto.ServiceStatus_SERVICE_STATUS_READY,
				ServerVersion: "1.0.0",
			},
		},
		{
			name: "VTS client error",
			mockSetup: func(m *MockVTSClient) {
				m.getServiceStateRes = nil
				m.getServiceStateErr = errors.New("VTS unavailable")
			},
			expectedResult: nil,
			expectedError:  "VTS unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockVTSClient(ctrl)
			tt.mockSetup(mockClient)
			
			provisioner := &Provisioner{VTSClient: mockClient}
			
			result, err := provisioner.GetVTSState()
			
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