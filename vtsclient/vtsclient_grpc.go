// Copyright 2022-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package vtsclient

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"
	"github.com/veraison/services/api"
	"github.com/veraison/services/config"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/vts/trustedservices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	ServerAddress string
	Credentials   credentials.TransportCredentials
	Connection    *grpc.ClientConn
}

// NewGRPC instantiate a new gRPC VTS client with default settings
func NewGRPC() *GRPC {
	return &GRPC{}
}

func (o *GRPC) Init(v *viper.Viper, certPath, keyPath string) error {
	cfg := trustedservices.NewGRPCConfig()

	loader := config.NewLoader(cfg)
	if err := loader.LoadFromViper(v); err != nil {
		return err
	}

	o.ServerAddress = cfg.ServerAddress

	if cfg.UseTLS {
		creds, err := trustedservices.LoadTLSCreds(certPath, keyPath, cfg.CACerts)
		if err != nil {
			return err
		}

		o.Credentials = creds
	} else {
		o.Credentials = insecure.NewCredentials()
	}

	return nil
}

func (o *GRPC) GetServiceState(
	ctx context.Context,
	in *emptypb.Empty,
	opts ...grpc.CallOption,
) (*proto.ServiceState, error) {
	if err := o.EnsureConnection(); err != nil {
		return nil, err
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

	mts, err := c.GetSupportedVerificationMediaTypes(ctx, in, opts...)
	if err != nil {
		return nil, err
	}

	return normalizeMediaTypeList(mts), nil
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

	mts, err := c.GetSupportedProvisioningMediaTypes(ctx, in, opts...)
	if err != nil {
		return nil, err
	}

	return normalizeMediaTypeList(mts), nil
}

func (o *GRPC) SubmitEndorsements(
	ctx context.Context, in *proto.SubmitEndorsementsRequest, opts ...grpc.CallOption,
) (*proto.SubmitEndorsementsResponse, error) {
	if err := o.EnsureConnection(); err != nil {
		return nil, NewNoConnectionError("SubmitEndorsements", err)
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

	conn, err := grpc.NewClient(o.ServerAddress, grpc.WithTransportCredentials(o.Credentials))
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

func (o *GRPC) GetEndorsements(
	ctx context.Context, in *proto.EndorsementQueryIn, opts ...grpc.CallOption,
) (*proto.EndorsementQueryOut, error) {
	if err := o.EnsureConnection(); err != nil {
		return nil, NewNoConnectionError("GetEndorsements", err)
	}

	c := o.GetProvisionerClient()
	if c == nil {
		return nil, ErrNoClient
	}

	return c.GetEndorsements(ctx, in, opts...)
}

func normalizeMediaTypeList(mts *proto.MediaTypeList) *proto.MediaTypeList {
	var nmts []string // nolint:prealloc

	for _, mt := range mts.GetMediaTypes() {
		nmt, err := api.NormalizeMediaType(mt)
		if err != nil {
			// skip invalid media type
			continue
		}
		nmts = append(nmts, nmt)
	}

	return &proto.MediaTypeList{MediaTypes: nmts}
}
