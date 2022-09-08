// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package trustedservices

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/veraison/services/config"
	"github.com/veraison/services/kvstore"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme"
	"github.com/veraison/services/vts/pluginmanager"
	"github.com/veraison/services/vts/policymanager"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

// XXX
// should be (also) serviceID
// should be passed as a parameter
const DummyTenantID = "0"

type GRPC struct {
	Config config.Store

	TaStore       kvstore.IKVStore
	EnStore       kvstore.IKVStore
	PluginManager pluginmanager.ISchemePluginManager
	PolicyManager *policymanager.PolicyManager

	Server *grpc.Server
	Socket net.Listener

	proto.UnimplementedVTSServer
}

func NewGRPC(
	cfg config.Store,
	taStore, enStore kvstore.IKVStore,
	pluginManager pluginmanager.ISchemePluginManager,
	policyManager *policymanager.PolicyManager,
) ITrustedServices {
	return &GRPC{
		Config:        cfg,
		TaStore:       taStore,
		EnStore:       enStore,
		PluginManager: pluginManager,
		PolicyManager: policyManager,
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

	// TODO load from config credentials for securing the transport endpoint
	var opts []grpc.ServerOption

	server := grpc.NewServer(opts...)
	proto.RegisterVTSServer(server, o)

	o.Socket = lsd
	o.Server = server

	return nil
}

func (o *GRPC) Close() error {
	if o.Server != nil {
		o.Server.GracefulStop()
	}

	if err := o.PluginManager.Close(); err != nil {
		log.Printf("plugin manager shutdown failed: %v", err)
	}

	if err := o.TaStore.Close(); err != nil {
		log.Printf("trust anchor store closure failed: %v", err)
	}

	if err := o.EnStore.Close(); err != nil {
		log.Printf("endorsement store closure failed: %v", err)
	}

	return nil
}

func (o *GRPC) AddSwComponents(ctx context.Context, req *proto.AddSwComponentsRequest) (*proto.AddSwComponentsResponse, error) {
	var (
		err    error
		keys   []string
		scheme scheme.IScheme
		val    []byte
	)

	for _, swComp := range req.GetSwComponents() {
		scheme, err = o.PluginManager.LookupByAttestationFormat(swComp.GetScheme())
		if err != nil {
			return addSwComponentErrorResponse(err), nil
		}

		keys, err = scheme.SynthKeysFromSwComponent(DummyTenantID, swComp)
		if err != nil {
			return addSwComponentErrorResponse(err), nil
		}

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
		err    error
		keys   []string
		scheme scheme.IScheme
		ta     *proto.Endorsement
		val    []byte
	)

	if req.TrustAnchor == nil {
		return addTrustAnchorErrorResponse(errors.New("nil trust anchor in request")), nil
	}

	ta = req.TrustAnchor

	scheme, err = o.PluginManager.LookupByAttestationFormat(ta.GetScheme())
	if err != nil {
		return addTrustAnchorErrorResponse(err), nil
	}

	keys, err = scheme.SynthKeysFromTrustAnchor(DummyTenantID, ta)
	if err != nil {
		return addTrustAnchorErrorResponse(err), nil
	}

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

func (o *GRPC) GetAttestation(ctx context.Context, token *proto.AttestationToken) (*proto.AppraisalContext, error) {
	scheme, err := o.PluginManager.LookupByMediaType(token.MediaType)
	if err != nil {
		return nil, err
	}

	ec, err := o.extractEvidence(scheme, token)
	if err != nil {
		return nil, err
	}

	endorsements, err := o.EnStore.Get(ec.SoftwareId)
	if err != nil && !errors.Is(err, kvstore.ErrKeyNotFound) {
		return nil, err
	}

	attestContext, err := scheme.AppraiseEvidence(ec, endorsements)
	if err != nil {
		return nil, err
	}

	err = o.PolicyManager.Evaluate(ctx, attestContext, endorsements)
	return attestContext, err
}

func (c *GRPC) extractEvidence(scheme scheme.IScheme, token *proto.AttestationToken) (*proto.EvidenceContext, error) {
	var err error

	ec := &proto.EvidenceContext{
		TenantId: token.TenantId,
		Format:   token.Format,
	}

	ec.TrustAnchorId, err = scheme.GetTrustAnchorID(token)
	if err != nil {
		return nil, err
	}

	trustAnchor, err := c.TaStore.Get(ec.TrustAnchorId)
	if err != nil {
		return nil, err
	}

	if len(trustAnchor) != 1 {
		return nil, fmt.Errorf("found %d trust anchors, want 1", len(trustAnchor))
	}

	extracted, err := scheme.ExtractVerifiedClaims(token, trustAnchor[0])
	if err != nil {
		return nil, err
	}

	ec.Evidence, err = structpb.NewStruct(extracted.ClaimsSet)
	if err != nil {
		return nil, err
	}

	ec.SoftwareId = extracted.SoftwareID

	return ec, nil
}

func (c *GRPC) GetSupportedVerificationMediaTypes(context.Context, *emptypb.Empty) (*proto.MediaTypeList, error) {
	mts, err := c.PluginManager.SupportedVerificationMediaTypes()
	if err != nil {
		return nil, fmt.Errorf("retrieving supported media types: %w", err)
	}

	return &proto.MediaTypeList{MediaTypes: mts}, nil
}
