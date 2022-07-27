// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package trustedservices

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/veraison/services/config"
	"github.com/veraison/services/kvstore"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/vts/pluginmanager"
	"google.golang.org/grpc"
)

type GRPC struct {
	Config config.Store

	TaStore       kvstore.IKVStore
	EnStore       kvstore.IKVStore
	PluginManager pluginmanager.IPluginManager

	Server *grpc.Server
	Socket net.Listener

	proto.UnimplementedVTSServer
}

func NewGRPC(cfg config.Store, taStore, enStore kvstore.IKVStore, pluginManager pluginmanager.IPluginManager) ITrustedServices {
	return &GRPC{
		Config:        cfg,
		TaStore:       taStore,
		EnStore:       enStore,
		PluginManager: pluginManager,
	}
}

func (o *GRPC) Run() error {
	if o.Server == nil {
		return errors.New("nil server: must call Init() first")
	}

	return o.Server.Serve(o.Socket)
}

func (o *GRPC) Init() error {
	addr, err := config.GetString(o.Config, "server.addr", &config.DefaultVTSAddr)
	if err != nil {
		return fmt.Errorf("loading configuration failed: %w", err)
	}

	lsd, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listening socket initialisation failed: %w", err)
	}

	// TODO credentials setup from config
	var opts []grpc.ServerOption

	server := grpc.NewServer(opts...)
	proto.RegisterVTSServer(server, o)

	o.Socket = lsd
	o.Server = server

	return nil
}

func (o *GRPC) Close() error {
	o.Server.GracefulStop()

	if err := o.PluginManager.Close(); err != nil {
		return fmt.Errorf("plugin manager termination failed: %w", err)
	}

	return nil
}

func (o *GRPC) AddSwComponents(ctx context.Context, req *proto.AddSwComponentsRequest) (*proto.AddSwComponentsResponse, error) {
	var (
		err  error
		keys []string
		//	scheme *common.SchemePlugin
		val []byte
	)

	for _, swComp := range req.GetSwComponents() {
		/*
			scheme, err = o.getSchemePlugin(swComp.GetScheme())
			if err != nil {
				return addSwComponentErrorResponse(err), nil
			}

			keys, err = scheme.SynthKeysFromSwComponent(DummyTenantID, swComp)
			if err != nil {
				return addSwComponentErrorResponse(err), nil
			}
		*/

		val, err = json.Marshal(swComp)
		if err != nil {
			return addSwComponentErrorResponse(err), nil
		}
	}

	for _, key := range keys {
		if err := o.EnStore.Add(key, string(val)); err != nil {
			if err != nil {
				return addSwComponentErrorResponse(err), nil
			}
		}
	}

	return addSwComponentSuccessResponse(), nil
}
func addSwComponentSuccessResponse() *proto.AddSwComponentsResponse {
	return &proto.AddSwComponentsResponse{
		Status: &proto.Status{
			Result: true,
		},
	}
}

func addSwComponentErrorResponse(err error) *proto.AddSwComponentsResponse {
	return &proto.AddSwComponentsResponse{
		Status: &proto.Status{
			Result:      false,
			ErrorDetail: fmt.Sprintf("%v", err),
		},
	}
}

func (o *GRPC) AddTrustAnchor(ctx context.Context, req *proto.AddTrustAnchorRequest) (*proto.AddTrustAnchorResponse, error) {
	var (
		err  error
		keys []string
		//		scheme *common.SchemePlugin
		ta  *proto.Endorsement
		val []byte
	)

	if req.TrustAnchor == nil {
		return addTrustAnchorErrorResponse(
			errors.New("nil TrustAnchor in request"),
		), nil
	}

	ta = req.TrustAnchor

	/*
		scheme, err = o.getSchemePlugin(ta.GetScheme())
		if err != nil {
			return addTrustAnchorErrorResponse(err), nil
		}

		keys, err = scheme.SynthKeysFromTrustAnchor(DummyTenantID, ta)
		if err != nil {
			return addTrustAnchorErrorResponse(err), nil
		}
	*/

	val, err = json.Marshal(ta)
	if err != nil {
		return addTrustAnchorErrorResponse(err), nil
	}

	for _, key := range keys {
		if err := o.TaStore.Add(key, string(val)); err != nil {
			if err != nil {
				return addTrustAnchorErrorResponse(err), nil
			}
		}
	}

	return addTrustAnchorSuccessResponse(), nil
}

func addTrustAnchorSuccessResponse() *proto.AddTrustAnchorResponse {
	return &proto.AddTrustAnchorResponse{
		Status: &proto.Status{
			Result: true,
		},
	}
}

func addTrustAnchorErrorResponse(err error) *proto.AddTrustAnchorResponse {
	return &proto.AddTrustAnchorResponse{
		Status: &proto.Status{
			Result:      false,
			ErrorDetail: fmt.Sprintf("%v", err),
		},
	}
}

func (o *GRPC) GetAttestation(ctx context.Context, token *proto.AttestationToken) (*proto.Attestation, error) {
	/*
		scheme, err := o.getSchemePlugin(token.Format)
		if err != nil {
			return nil, err
		}

		ec, err := o.extractEvidence(scheme, token)
		if err != nil {
			return nil, err
		}

		endorsements, err := o.EnStore.Get(ec.SoftwareId)
		if err != nil {
			return nil, err
		}

		return scheme.GetAttestation(ec, endorsements)
	*/
	return nil, nil
}
