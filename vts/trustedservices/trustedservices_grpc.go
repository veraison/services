// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package trustedservices

import (
	"context"
	"encoding/base64"
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
	"github.com/veraison/services/handler"
	"github.com/veraison/services/kvstore"
	"github.com/veraison/services/plugin"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/vts/appraisal"
	"github.com/veraison/services/vts/earsigner"
	"github.com/veraison/services/vts/policymanager"
)

// XXX
// should be (also) serviceID
// should be passed as a parameter
const DummyTenantID = "0"

// Supported parameters:
//
//   - vts.server-addr: string w/ syntax specified in
//     https://github.com/grpc/grpc/blob/master/doc/naming.md
//
//   - TODO(tho) load balancing config
//     See https://github.com/grpc/grpc/blob/master/doc/load-balancing.md
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
	PluginManager plugin.IManager[handler.IEvidenceHandler]
	PolicyManager *policymanager.PolicyManager
	EarSigner     earsigner.IEarSigner

	Server *grpc.Server
	Socket net.Listener

	logger *zap.SugaredLogger

	proto.UnimplementedVTSServer
}

func NewGRPC(
	taStore, enStore kvstore.IKVStore,
	pluginManager plugin.IManager[handler.IEvidenceHandler],
	policyManager *policymanager.PolicyManager,
	earSigner earsigner.IEarSigner,
	logger *zap.SugaredLogger,
) ITrustedServices {
	return &GRPC{
		TaStore:       taStore,
		EnStore:       enStore,
		PluginManager: pluginManager,
		PolicyManager: policyManager,
		EarSigner:     earSigner,
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

func (o *GRPC) Init(v *viper.Viper, pm plugin.IManager[handler.IEvidenceHandler]) error {
	var err error

	cfg := GRPCConfig{ServerAddress: DefaultVTSAddr}

	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromViper(v); err != nil {
		return err
	}

	o.PluginManager = pm

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

	if err := o.EarSigner.Close(); err != nil {
		o.logger.Errorf("EAR signer closure failed: %v", err)
	}

	return nil
}

func (o *GRPC) GetServiceState(context.Context, *emptypb.Empty) (*proto.ServiceState, error) {
	mediaTypes := o.PluginManager.GetRegisteredMediaTypes()

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

func (o *GRPC) AddRefValues(ctx context.Context, req *proto.AddRefValuesRequest) (*proto.AddRefValuesResponse, error) {
	var (
		err     error
		keys    []string
		handler handler.IEvidenceHandler
		val     []byte
	)

	o.logger.Debugw("AddRefValue", "ref-value", req.ReferenceValues)

	for _, refVal := range req.GetReferenceValues() {
		handler, err = o.PluginManager.LookupByAttestationScheme(refVal.GetScheme())
		if err != nil {
			return addRefValueErrorResponse(err), nil
		}

		keys, err = handler.SynthKeysFromRefValue(DummyTenantID, refVal)
		if err != nil {
			return addRefValueErrorResponse(err), nil
		}

		val, err = json.Marshal(refVal)
		if err != nil {
			return addRefValueErrorResponse(err), nil
		}
	}

	for _, key := range keys {
		if err := o.EnStore.Add(key, string(val)); err != nil {
			if err != nil {
				return addRefValueErrorResponse(err), nil
			}
		}
	}

	o.logger.Infow("added reference values", "keys", keys)

	return addRefValueSuccessResponse(), nil
}
func addRefValueSuccessResponse() *proto.AddRefValuesResponse {
	return &proto.AddRefValuesResponse{
		Status: &proto.Status{
			Result: true,
		},
	}
}

func addRefValueErrorResponse(err error) *proto.AddRefValuesResponse {
	return &proto.AddRefValuesResponse{
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
		err     error
		keys    []string
		handler handler.IEvidenceHandler
		ta      *proto.Endorsement
		val     []byte
	)

	o.logger.Debugw("AddTrustAnchor", "trust-anchor", req.TrustAnchor)

	if req.TrustAnchor == nil {
		return addTrustAnchorErrorResponse(errors.New("nil trust anchor in request")), nil
	}

	ta = req.TrustAnchor

	handler, err = o.PluginManager.LookupByAttestationScheme(ta.GetScheme())
	if err != nil {
		return addTrustAnchorErrorResponse(err), nil
	}

	keys, err = handler.SynthKeysFromTrustAnchor(DummyTenantID, ta)
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
		"tenant-id", token.TenantId)

	handler, err := o.PluginManager.LookupByMediaType(token.MediaType)
	if err != nil {
		return nil, err
	}

	appraisal, err := o.initEvidenceContext(handler, token)
	if err != nil {
		return nil, err
	}

	ta, err := o.getTrustAnchor(appraisal.EvidenceContext.TrustAnchorId)
	if err != nil {
		return nil, err
	}

	extracted, err := handler.ExtractClaims(token, ta)
	if err != nil {
		return nil, err
	}

	appraisal.EvidenceContext.Evidence, err = structpb.NewStruct(extracted.ClaimsSet)
	if err != nil {
		return nil, err
	}

	appraisal.EvidenceContext.ReferenceId = extracted.ReferenceID

	o.logger.Debugw("constructed evidence context",
		"software-id", appraisal.EvidenceContext.ReferenceId,
		"trust-anchor-id", appraisal.EvidenceContext.TrustAnchorId)

	endorsements, err := o.EnStore.Get(appraisal.EvidenceContext.ReferenceId)
	if err != nil && !errors.Is(err, kvstore.ErrKeyNotFound) {
		return nil, err
	}

	if len(endorsements) > 0 {
		o.logger.Debugw("obtained endorsements", "endorsements", endorsements)
	}

	if err = handler.ValidateEvidenceIntegrity(token, ta, endorsements); err != nil {
		// TODO(setrofim): we should distinguish between validation
		// failing due to bad signature vs actual error here, and only
		// return actual err. Bad sig should be reported as a failure
		// in attestation result, rather than an error in the
		// attestation call.
		return nil, err
	}

	// TODO(tho) we need to clearly define what it means for a plugin to return
	// an error and decide whether / how such condition gets mapped into the
	// AR4SI trust vector (ISTM that a VerifierMalfunctionClaim (-1) may be the
	// right signal.)
	appraisedResult, err := handler.AppraiseEvidence(appraisal.EvidenceContext, endorsements)
	if err != nil {
		return nil, err
	}
	appraisal.Result = appraisedResult

	// TODO(setrofim) Should we be doing appraisal.SetError() on error here?
	// This should be decided as part of a wider policy framework design.
	err = o.PolicyManager.Evaluate(ctx, appraisal, endorsements)
	if err != nil {
		return nil, err
	}

	appraisal.Result.UpdateStatusFromTrustVector()

	ear, err := o.EarSigner.Sign(*appraisal.Result)
	if err != nil {
		return nil, err
	}
	appraisal.SignedEAR = ear

	appraisal.Result.VerifierID.Build = &config.Version
	appraisal.Result.VerifierID.Developer = &config.Developer

	encodedNonce := base64.URLEncoding.EncodeToString(token.Nonce)
	appraisal.Result.Nonce = &encodedNonce

	o.logger.Infow("evaluated attestation result", "attestation-result", appraisal.Result)

	return appraisal.GetContext(), err
}

func (c *GRPC) initEvidenceContext(
	handler handler.IEvidenceHandler,
	token *proto.AttestationToken,
) (*appraisal.Appraisal, error) {
	var err error

	appraisal := appraisal.New(token.TenantId, handler.GetAttestationScheme())

	appraisal.EvidenceContext.TrustAnchorId, err = handler.GetTrustAnchorID(token)
	if err != nil {
		return nil, err
	}

	return appraisal, nil
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
	mts := c.PluginManager.GetRegisteredMediaTypes()
	return &proto.MediaTypeList{MediaTypes: mts}, nil
}

func (o *GRPC) GetEARSigningPublicKey(context.Context, *emptypb.Empty) (*proto.PublicKey, error) {
	key, err := o.EarSigner.GetEARSigningPublicKeyEar()
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(key)
	if err != nil {
		return nil, err
	}

	bstring := string(b)

	return &proto.PublicKey{
		Key: bstring,
	}, nil
}
