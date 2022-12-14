// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package trustedservices

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/veraison/services/config"
	"github.com/veraison/services/kvstore"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme"
	"github.com/veraison/services/vts/pluginmanager"
	"github.com/veraison/services/vts/policymanager"
)

// XXX
// should be (also) serviceID
// should be passed as a parameter
const DummyTenantID = "0"

// Supported parameters:
// * vts.server-addr: string w/ syntax specified in
//   https://github.com/grpc/grpc/blob/master/doc/naming.md
//
// * TODO(tho) load balancing config
//   See https://github.com/grpc/grpc/blob/master/doc/load-balancing.md
//
// * TODO(tho) auth'n credentials (e.g., TLS / JWT credentials)
type GRPCConfig struct {
	ServerAddress string `mapstructure:"server-addr" valid:"dialstring"`
	ListenAddress string `mapstructure:"listen-addr" valid:"dialstring" config:"zerodefault"`
}

func NewGRPCConfig() *GRPCConfig {
	return &GRPCConfig{ServerAddress: DefaultVTSAddr}
}

type GRPC struct {
	ServerAddress string

	TaStore       kvstore.IKVStore
	EnStore       kvstore.IKVStore
	PluginManager pluginmanager.ISchemePluginManager
	PolicyManager *policymanager.PolicyManager

	Server *grpc.Server
	Socket net.Listener

	logger *zap.SugaredLogger

	proto.UnimplementedVTSServer
}

func NewGRPC(
	taStore, enStore kvstore.IKVStore,
	pluginManager pluginmanager.ISchemePluginManager,
	policyManager *policymanager.PolicyManager,
	logger *zap.SugaredLogger,
) ITrustedServices {
	return &GRPC{
		TaStore:       taStore,
		EnStore:       enStore,
		PluginManager: pluginManager,
		PolicyManager: policyManager,
		logger:        logger,
	}
}

func (o *GRPC) Run() error {
	if o.Server == nil {
		return errors.New("nil server: must call Init() first")
	}

	o.logger.Infow("listening for GRPC requests", "address", o.ServerAddress)
	return o.Server.Serve(o.Socket)
}

func (o *GRPC) Init(v *viper.Viper) error {
	cfg := GRPCConfig{ServerAddress: DefaultVTSAddr}

	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromViper(v); err != nil {
		return err
	}

	if cfg.ListenAddress != "" {
		o.ServerAddress = cfg.ListenAddress
	} else {
		// note: the indexing will succeed as ServerAddress has been validated as a dialstring.
		o.ServerAddress = ":" + strings.Split(cfg.ServerAddress, ":")[1]
	}

	lsd, err := net.Listen("tcp", o.ServerAddress)
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
		o.logger.Errorf("plugin manager shutdown failed: %v", err)
	}

	if err := o.TaStore.Close(); err != nil {
		o.logger.Errorf("trust anchor store closure failed: %v", err)
	}

	if err := o.EnStore.Close(); err != nil {
		o.logger.Errorf("endorsement store closure failed: %v", err)
	}

	return nil
}

func (o *GRPC) GetServiceState(context.Context, *emptypb.Empty) (*proto.ServiceState, error) {

	mediaTypes, err := o.PluginManager.SupportedVerificationMediaTypes()
	if err != nil {
		return nil, err
	}

	mediaTypesList, err := proto.NewStringList(mediaTypes)
	if err != nil {
		return nil, err
	}

	return &proto.ServiceState{
		Status:        proto.ServiceStatus_READY,
		ServerVersion: config.Version,
		SupportedMediaTypes: map[string]*structpb.ListValue{
			"challenge-response/v1": mediaTypesList.AsListValue(),
		},
	}, nil
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

	o.logger.Infow("added software component", "keys", keys)

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

func (o *GRPC) AddTrustAnchor(
	ctx context.Context,
	req *proto.AddTrustAnchorRequest,
) (*proto.AddTrustAnchorResponse, error) {
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

	o.logger.Infow("added trust anchor", "keys", keys)

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

func (o *GRPC) GetAttestation(
	ctx context.Context,
	token *proto.AttestationToken,
) (*proto.AppraisalContext, error) {
	o.logger.Infow("get attestation", "media-type", token.MediaType,
		"tenant-id", token.TenantId, "format", token.Format)

	scheme, err := o.PluginManager.LookupByMediaType(token.MediaType)
	if err != nil {
		return nil, err
	}

	ec, err := o.initEvidenceContext(scheme, token)
	if err != nil {
		return nil, err
	}

	ta, err := o.getTrustAnchor(ec.TrustAnchorId)
	if err != nil {
		return nil, err
	}

	extracted, err := scheme.ExtractClaims(token, ta)
	if err != nil {
		return nil, err
	}

	ec.Evidence, err = structpb.NewStruct(extracted.ClaimsSet)
	if err != nil {
		return nil, err
	}

	ec.SoftwareId = extracted.SoftwareID

	o.logger.Debugw("constructed evidence context", "software-id", ec.SoftwareId,
		"trust-anchor-id", ec.TrustAnchorId)

	var endorsements []string
	if ec.SoftwareId != "" {
		endorsements, err = o.EnStore.Get(ec.SoftwareId)
		if err != nil && !errors.Is(err, kvstore.ErrKeyNotFound) {

			return nil, err
		}
	}

	if len(endorsements) > 0 {
		o.logger.Debugw("obtained endorsements", "endorsements", endorsements)
	}

	if err = scheme.ValidateEvidenceIntegrity(token, ta, endorsements); err != nil {
		// TODO(setrofim): we should distinguish between validation
		// failing due to bad signature vs actual error here, and only
		// return actual err. Bad sig should be reported as a failure
		// in attestation result, rather than an error in the
		// attestation call.
		return nil, err
	}

	attestContext, err := scheme.AppraiseEvidence(ec, endorsements)
	if err != nil {
		attestContext.Result.SetVerifierError()
		return nil, err
	}

	// TODO(setrofim) Should we be doing SetVerifierError() on error here?
	// This should be diced as part of wider policy framework desing.
	err = o.PolicyManager.Evaluate(ctx, attestContext, endorsements)

	o.logger.Infow("evaluated attestation result", "attestation-result", attestContext.Result)

	return attestContext, err
}

func (c *GRPC) initEvidenceContext(
	scheme scheme.IScheme,
	token *proto.AttestationToken,
) (*proto.EvidenceContext, error) {
	var err error

	ec := &proto.EvidenceContext{
		TenantId: token.TenantId,
		Format:   token.Format,
	}

	ec.TrustAnchorId, err = scheme.GetTrustAnchorID(token)
	if err != nil {
		return nil, err
	}

	return ec, nil
}

func (c *GRPC) getTrustAnchor(id string) (string, error) {
	values, err := c.TaStore.Get(id)
	if err != nil {
		return "", err
	}

	if len(values) != 1 {
		return "", fmt.Errorf("found %d trust anchors, want 1", len(values))
	}

	return values[0], nil
}

func (c *GRPC) GetSupportedVerificationMediaTypes(context.Context, *emptypb.Empty) (*proto.MediaTypeList, error) {
	mts, err := c.PluginManager.SupportedVerificationMediaTypes()
	if err != nil {
		return nil, fmt.Errorf("retrieving supported media types: %w", err)
	}

	return &proto.MediaTypeList{MediaTypes: mts}, nil
}
