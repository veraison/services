// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"errors"
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/veraison/services/pkg/config"
	"github.com/veraison/services/pkg/endorsement"
	"github.com/veraison/services/pkg/fault"
	"github.com/veraison/services/pkg/policy"
	"github.com/veraison/services/pkg/scheme"
	"github.com/veraison/services/pkg/util"
	"github.com/veraison/services/pkg/vts"
)

var ErrMeasurementsNotSupported = errors.New("measurements in CoSERV queries are not supported")

type Server struct {
	config  *vts.ServerConfig
	service vts.IService
	server  *grpc.Server
	socket  net.Listener
	logger  *zap.SugaredLogger

	vts.UnimplementedVTSServer
}

func CreateServer(
	cfg *vts.ServerConfig,
	endorsementStore *endorsement.Store,
	schemeManager *scheme.Manager,
	policyManager *policy.Manager,
	earSigner IEarSigner,
	logger *zap.SugaredLogger,
) (*Server, error) {
	socket, err := net.Listen("tcp", cfg.ListenAddress)
	if err != nil {
		return nil, fmt.Errorf("listening socket initialisation failed: %w", err)
	}

	server := &Server{
		config: cfg,
		service: NewService(
			endorsementStore,
			schemeManager,
			policyManager,
			earSigner,
			cfg.CoRIMCACerts,
			logger.Named("service"),
		),
		socket: socket,
		logger: logger,
	}

	var opts []grpc.ServerOption
	if cfg.TLSCreds != nil {
		opts = append(opts, grpc.Creds(cfg.TLSCreds))
	}

	grpcServer := grpc.NewServer(opts...)
	vts.RegisterVTSServer(grpcServer, server)
	server.server = grpcServer

	return server, nil
}

func (o *Server) Run() error {
	o.logger.Infof("listening for GRPC requests on %s", o.config.ListenAddress)
	return o.server.Serve(o.socket)
}

func (o *Server) Close() {
	o.server.GracefulStop()
}

func (o *Server) GetServiceState(
	ctx context.Context,
	notused *emptypb.Empty,
) (*vts.GetServiceStateResponse, error) {
	state, err := o.service.GetServiceState(ctx)
	if err != nil {
		return nil, err
	}

	provisioningMediaTypes, err := util.SliceToListValue(state.ProvisioningMediaTypes)
	if err != nil {
		return nil, err
	}

	verificationMediaTypes, err := util.SliceToListValue(state.VerificationMediaTypes)
	if err != nil {
		return nil, err
	}

	coservMediaTypes, err := util.SliceToListValue(state.CoservMediaTypes)
	if err != nil {
		return nil, err
	}

	return &vts.GetServiceStateResponse{
		ServerVersion:          config.Version,
		ProvisioningMediaTypes: provisioningMediaTypes,
		VerificationMediaTypes: verificationMediaTypes,
		CoservMediaTypes:       coservMediaTypes,
		EarSigningPublicKey:    state.EarSigningPublicKey,
		CoservSigningPublicKey: state.CoservSigningPublicKey,
	}, nil
}

func (o *Server) SubmitEndorsements(
	ctx context.Context,
	request *vts.SubmitEndorsementsRequest,
) (*emptypb.Empty, error) {
	err := o.service.SubmitEndorsements(ctx, request.MediaType, request.Data)
	return vts.Empty, fault.SerializeError(err)
}

func (o *Server) GetAttestation(
	ctx context.Context,
	request *vts.GetAttestationRequest,
) (*vts.GetAttestationResponse, error) {
	evidence := &scheme.Evidence{
		Data:      request.Data,
		MediaType: request.MediaType,
		Nonce:     request.Nonce,
	}

	resultBytes, err := o.service.GetAttestation(ctx, evidence)
	return &vts.GetAttestationResponse{SignedEar: resultBytes}, fault.SerializeError(err)

}

func (o *Server) GetEndorsements(
	ctx context.Context,
	request *vts.GetEndosementsRequest,
) (*vts.GetEndosementsResponse, error) {
	coservBytes, err := o.service.GetEndorsements(ctx, request.MediaType, request.Query)
	return &vts.GetEndosementsResponse{ResultSet: coservBytes}, fault.SerializeError(err)
}

func (o *Server) SetEndorsementsState(
	ctx context.Context,
	request *vts.SetEndorsementsStateRequest,
) (*emptypb.Empty, error) {
	err := o.service.SetEndorsementsState(ctx, request.Data, request.SetActive)
	return vts.Empty, fault.SerializeError(err)
}
