// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package vtsclient

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/veraison/services/config"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/vts/trustedservices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	ErrNoClient = errors.New("there is no active gRPC VTS client")
)

type NoConnectionError struct {
	Context string
	Err     error
}

func NewNoConnectionError(ctx string, err error) NoConnectionError {
	return NoConnectionError{Context: ctx, Err: err}
}

func (o NoConnectionError) Error() string {
	return fmt.Sprintf("(from: %s) %v", o.Context, o.Err)
}

func (o NoConnectionError) Unwrap() error {
	return o.Err
}

type GRPC struct {
	ServerAddress     string
	Connection        *grpc.ClientConn
	ConnectionTimeout time.Duration
}

// NewGRPC instantiate a new gRPC VTS client with default settings
func NewGRPC() *GRPC {
	return &GRPC{
		ConnectionTimeout: time.Second,
	}
}

func (o *GRPC) Init(v *viper.Viper) error {
	cfg := trustedservices.NewGRPCConfig()

	loader := config.NewLoader(cfg)
	if err := loader.LoadFromViper(v); err != nil {
		return err
	}

	o.ServerAddress = cfg.ServerAddress

	return nil
}

func (o *GRPC) GetServiceState(
	ctx context.Context,
	in *emptypb.Empty,
	opts ...grpc.CallOption,
) (*proto.ServiceState, error) {
	if err := o.EnsureConnection(); err != nil {
		return &proto.ServiceState{
			Status: proto.ServiceStatus_SERVICE_STATUS_DOWN,
		}, nil
	}

	c := o.GetProvisionerClient()
	if c == nil {
		return nil, ErrNoClient
	}

	return c.GetServiceState(ctx, in, opts...)
}

func (o *GRPC) GetAttestation(
	ctx context.Context, in *proto.AttestationToken, opts ...grpc.CallOption,
) (*proto.AppraisalContext, error) {
	if err := o.EnsureConnection(); err != nil {
		return nil, NewNoConnectionError("GetAttestation", err)
	}

	c := o.GetProvisionerClient()
	if c == nil {
		return nil, ErrNoClient
	}

	return c.GetAttestation(ctx, in, opts...)
}

func (o *GRPC) GetSupportedVerificationMediaTypes(
	ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption,
) (*proto.MediaTypeList, error) {
	if err := o.EnsureConnection(); err != nil {
		return nil, NewNoConnectionError("GetSupportedVerificationMediaTypes", err)
	}

	c := o.GetProvisionerClient()
	if c == nil {
		return nil, ErrNoClient
	}

	return c.GetSupportedVerificationMediaTypes(ctx, in, opts...)
}

func (o *GRPC) GetSupportedProvisioningMediaTypes(
	ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption,
) (*proto.MediaTypeList, error) {
	if err := o.EnsureConnection(); err != nil {
		return nil, NewNoConnectionError("GetSupportedProvisioningMediaTypes", err)
	}

	c := o.GetProvisionerClient()
	if c == nil {
		return nil, ErrNoClient
	}

	return c.GetSupportedProvisioningMediaTypes(ctx, in, opts...)
}

func (o *GRPC) SubmitEndorsements(
	ctx context.Context, in *proto.SubmitEndorsementsRequest, opts ...grpc.CallOption,
) (*proto.SubmitEndorsementsResponse, error) {
	if err := o.EnsureConnection(); err != nil {
		return nil, NewNoConnectionError("GetSupportedVerificationMediaTypes", err)
	}
	c := o.GetProvisionerClient()
	if c == nil {
		return nil, ErrNoClient
	}
	return c.SubmitEndorsements(ctx, in, opts...)
}

func (o *GRPC) GetProvisionerClient() proto.VTSClient {
	if o.Connection == nil {
		return nil
	}

	return proto.NewVTSClient(o.Connection)
}

func (o *GRPC) EnsureConnection() error {
	if o.Connection != nil {
		return nil
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), o.ConnectionTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, o.ServerAddress, opts...)
	if err != nil {
		return fmt.Errorf("connection to gRPC VTS server [%s] failed: %w", o.ServerAddress, err)
	}

	o.Connection = conn

	return nil
}

func (o *GRPC) GetEARSigningPublicKey(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*proto.PublicKey, error) {
	if err := o.EnsureConnection(); err != nil {
		return nil, NewNoConnectionError("GetEARSigningPublicKey", err)
	}

	c := o.GetProvisionerClient()
	if c == nil {
		return nil, ErrNoClient
	}

	return c.GetEARSigningPublicKey(ctx, in, opts...)
}
