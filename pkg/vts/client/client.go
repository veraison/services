// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package client

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/veraison/services/pkg/fault"
	"github.com/veraison/services/pkg/scheme"
	"github.com/veraison/services/pkg/util"
	"github.com/veraison/services/pkg/vts"
)

type TrustedService struct {
	config *vts.ClientConfig
	logger *zap.SugaredLogger
	opts   []grpc.CallOption

	connection *grpc.ClientConn
	client     vts.VTSClient
}

// NewGRPC instantiate a new gRPC VTS client with default settings
func NewTrustedService(
	config *vts.ClientConfig,
	logger *zap.SugaredLogger,
	opts ...grpc.CallOption,
) *TrustedService {
	return &TrustedService{
		config: config,
		logger: logger,
		opts:   opts,
	}
}

func (o *TrustedService) GetServiceState(
	ctx context.Context,
) (*vts.State, error) {
	if err := o.ensureConnection(); err != nil {
		return nil, fault.NoConnection(err)
	}

	response, err := o.client.GetServiceState(ctx, vts.Empty, o.opts...)
	if err != nil {
		return nil, fault.ParseError(err)
	}

	provisioningMediaTypes, err := util.ListValueToSlice[string](response.ProvisioningMediaTypes)
	if err != nil {
		return nil, err
	}

	verificationMediaTypes, err := util.ListValueToSlice[string](response.VerificationMediaTypes)
	if err != nil {
		return nil, err
	}

	coservMediaTypes, err := util.ListValueToSlice[string](response.CoservMediaTypes)
	if err != nil {
		return nil, err
	}

	return &vts.State{
		Version:                response.ServerVersion,
		ProvisioningMediaTypes: provisioningMediaTypes,
		VerificationMediaTypes: verificationMediaTypes,
		CoservMediaTypes:       coservMediaTypes,
		EarSigningPublicKey:    response.EarSigningPublicKey,
		CoservSigningPublicKey: response.CoservSigningPublicKey,
	}, nil
}

func (o *TrustedService) GetAttestation(
	ctx context.Context,
	evidence *scheme.Evidence,
) ([]byte, error) {
	if err := o.ensureConnection(); err != nil {
		return nil, fault.NoConnection(err)
	}

	request := &vts.GetAttestationRequest{
		Data:      evidence.Data,
		MediaType: evidence.MediaType,
		Nonce:     evidence.Nonce,
	}

	response, err := o.client.GetAttestation(ctx, request, o.opts...)
	if err != nil {
		return nil, fault.ParseError(err)
	}

	return response.SignedEar, nil
}

func (o *TrustedService) SubmitEndorsements(
	ctx context.Context,
	mediaType string,
	data []byte,
) error {
	if err := o.ensureConnection(); err != nil {
		return fault.NoConnection(err)
	}

	request := &vts.SubmitEndorsementsRequest{
		MediaType: mediaType,
		Data:      data,
	}

	_, err := o.client.SubmitEndorsements(ctx, request, o.opts...)
	return fault.ParseError(err)
}

func (o *TrustedService) SetEndorsementsState(
	ctx context.Context,
	data []byte,
	active bool,
) error {
	if err := o.ensureConnection(); err != nil {
		return fault.NoConnection(err)
	}

	request := &vts.SetEndorsementsStateRequest{
		Data:      data,
		SetActive: active,
	}

	_, err := o.client.SetEndorsementsState(ctx, request, o.opts...)
	return fault.ParseError(err)
}

func (o *TrustedService) GetEndorsements(
	ctx context.Context,
	mediaType string,
	query string,
) ([]byte, error) {
	if err := o.ensureConnection(); err != nil {
		return nil, fault.NoConnection(err)
	}

	request := &vts.GetEndosementsRequest{
		MediaType: mediaType,
		Query:     query,
	}

	response, err := o.client.GetEndorsements(ctx, request, o.opts...)
	if err != nil {
		return nil, fault.ParseError(err)
	}

	return response.ResultSet, nil
}

func (o *TrustedService) ensureConnection() error {
	if o.client == nil {
		return o.connect()
	}

	return nil
}

func (o *TrustedService) connect() error {
	if o.connection != nil {
		o.logger.Debug("closing existing connection")
		if err := o.connection.Close(); err != nil {
			return err
		}
	}

	conn, err := grpc.NewClient(
		o.config.ServerAddress,
		grpc.WithTransportCredentials(o.config.TLSCreds),
	)
	if err != nil {
		return fmt.Errorf(
			"connection to gRPC VTS server [%s] failed: %w",
			o.config.ServerAddress,
			err,
		)
	}

	o.connection = conn
	o.client = vts.NewVTSClient(o.connection)

	return nil
}
