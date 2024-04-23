// Copyright 2022-2024 Contributors to the Veraison project.
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

	"github.com/veraison/ear"
	"github.com/veraison/services/config"
	"github.com/veraison/services/handler"
	handlermod "github.com/veraison/services/handler"
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

	TaStore            kvstore.IKVStore
	EnStore            kvstore.IKVStore
	EvPluginManager    plugin.IManager[handler.IEvidenceHandler]
	EndPluginManager   plugin.IManager[handler.IEndorsementHandler]
	StorePluginManager plugin.IManager[handler.IStoreHandler]
	PolicyManager      *policymanager.PolicyManager
	EarSigner          earsigner.IEarSigner

	Server *grpc.Server
	Socket net.Listener

	logger *zap.SugaredLogger

	proto.UnimplementedVTSServer
}

func NewGRPC(
	taStore, enStore kvstore.IKVStore,
	evpluginManager plugin.IManager[handler.IEvidenceHandler],
	endpluginManager plugin.IManager[handler.IEndorsementHandler],
	storepluginManager plugin.IManager[handler.IStoreHandler],
	policyManager *policymanager.PolicyManager,
	earSigner earsigner.IEarSigner,
	logger *zap.SugaredLogger,
) ITrustedServices {
	return &GRPC{
		TaStore:            taStore,
		EnStore:            enStore,
		EvPluginManager:    evpluginManager,
		EndPluginManager:   endpluginManager,
		StorePluginManager: storepluginManager,
		PolicyManager:      policyManager,
		EarSigner:          earSigner,
		logger:             logger,
	}
}

func (o *GRPC) Run() error {
	if o.Server == nil {
		return errors.New("nil server: must call Init() first")
	}

	o.logger.Infow("listening for GRPC requests", "address", o.ServerAddress)
	return o.Server.Serve(o.Socket)
}

func (o *GRPC) Init(
	v *viper.Viper,
	evm plugin.IManager[handler.IEvidenceHandler],
	endm plugin.IManager[handler.IEndorsementHandler],
	stm plugin.IManager[handler.IStoreHandler],
) error {
	var err error

	cfg := GRPCConfig{ServerAddress: DefaultVTSAddr}

	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromViper(v); err != nil {
		return err
	}

	o.EvPluginManager = evm
	o.EndPluginManager = endm
	o.StorePluginManager = stm

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

	if err := o.EvPluginManager.Close(); err != nil {
		o.logger.Errorf("evidence plugin manager shutdown failed: %v", err)
	}

	if err := o.EndPluginManager.Close(); err != nil {
		o.logger.Errorf("endorsement plugin manager shutdown failed: %v", err)
	}

	if err := o.StorePluginManager.Close(); err != nil {
		o.logger.Errorf("store plugin manager shutdown failed: %v", err)
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
	mediaTypes := o.EvPluginManager.GetRegisteredMediaTypes()

	mediaTypesList, err := proto.NewStringList(mediaTypes)
	if err != nil {
		return nil, err
	}

	return &proto.ServiceState{
		Status:        proto.ServiceStatus_SERVICE_STATUS_READY,
		ServerVersion: config.Version,
		SupportedMediaTypes: map[string]*structpb.ListValue{
			"challenge-response/v1": mediaTypesList.AsListValue(),
		},
	}, nil
}

func (o *GRPC) SubmitEndorsements(ctx context.Context, req *proto.SubmitEndorsementsRequest) (*proto.SubmitEndorsementsResponse, error) {
	o.logger.Debugw("SubmitEndorsements", "media-type", req.MediaType)

	handlerPlugin, err := o.EndPluginManager.LookupByMediaType(req.MediaType)
	if err != nil {
		return nil, err
	}

	rsp, err := handlerPlugin.Decode(req.Data)
	if err != nil {
		return submitEndorsementErrorResponse(err), nil
	}
	if err := o.storeEndorsements(ctx, rsp); err != nil {
		return submitEndorsementErrorResponse(err), nil
	}
	return submitEndorsementSuccessResponse(), nil
}

func (o *GRPC) storeEndorsements(ctx context.Context, rsp *handler.EndorsementHandlerResponse) error {
	for _, ta := range rsp.TrustAnchors {

		err := o.addTrustAnchor(ctx, &ta)
		if err != nil {
			return fmt.Errorf("store operation failed for trust anchor: %w", err)
		}
	}

	for _, refVal := range rsp.ReferenceValues {

		err := o.addRefValues(ctx, &refVal)
		if err != nil {
			return fmt.Errorf("store operation failed for reference values: %w", err)
		}
	}

	return nil
}

func submitEndorsementSuccessResponse() *proto.SubmitEndorsementsResponse {
	return &proto.SubmitEndorsementsResponse{
		Status: &proto.Status{
			Result: true,
		},
	}
}

func submitEndorsementErrorResponse(err error) *proto.SubmitEndorsementsResponse {
	return &proto.SubmitEndorsementsResponse{
		Status: &proto.Status{
			Result:      false,
			ErrorDetail: fmt.Sprintf("%v", err),
		},
	}
}

func (o *GRPC) addRefValues(ctx context.Context, refVal *handler.Endorsement) error {
	var (
		err     error
		keys    []string
		handler handler.IStoreHandler
		val     []byte
	)

	handler, err = o.StorePluginManager.LookupByAttestationScheme(refVal.Scheme)
	if err != nil {
		return err
	}

	keys, err = handler.SynthKeysFromRefValue(DummyTenantID, refVal)
	if err != nil {
		return err
	}

	val, err = json.Marshal(refVal)
	if err != nil {
		return err
	}

	for _, key := range keys {
		if err := o.EnStore.Add(key, string(val)); err != nil {
			if err != nil {
				return err
			}
		}
	}

	o.logger.Infow("added reference values", "keys", keys)

	return nil
}

func (o *GRPC) addTrustAnchor(
	ctx context.Context,
	req *handler.Endorsement,
) error {
	var (
		err     error
		keys    []string
		handler handler.IStoreHandler
		val     []byte
	)

	o.logger.Debugw("AddTrustAnchor", "trust-anchor", req)

	if req == nil {
		return errors.New("nil trust anchor in request")
	}

	handler, err = o.StorePluginManager.LookupByAttestationScheme(req.Scheme)
	if err != nil {
		return err
	}

	keys, err = handler.SynthKeysFromTrustAnchor(DummyTenantID, req)
	if err != nil {
		return err
	}

	val, err = json.Marshal(req)
	if err != nil {
		return err
	}

	for _, key := range keys {
		if err := o.TaStore.Add(key, string(val)); err != nil {
			if err != nil {
				return err
			}
		}
	}

	o.logger.Infow("added trust anchor", "keys", keys)

	return nil
}

func (o *GRPC) GetAttestation(
	ctx context.Context,
	token *proto.AttestationToken,
) (*proto.AppraisalContext, error) {
	o.logger.Infow("get attestation", "media-type", token.MediaType,
		"tenant-id", token.TenantId)

	handler, err := o.EvPluginManager.LookupByMediaType(token.MediaType)
	if err != nil {
		appraisal := appraisal.New(token.TenantId, token.Nonce, "ERROR")
		appraisal.SetAllClaims(ear.UnexpectedEvidenceClaim)
		appraisal.AddPolicyClaim("problem", "could not resolve media type")
		return o.finalize(appraisal, err)
	}

	scheme := handler.GetAttestationScheme()
	stHandler, err := o.StorePluginManager.LookupByAttestationScheme(scheme)
	if err != nil {
		appraisal := appraisal.New(token.TenantId, token.Nonce, "ERROR")
		appraisal.SetAllClaims(ear.UnexpectedEvidenceClaim)
		appraisal.AddPolicyClaim("problem", "could not resolve scheme name")
		return o.finalize(appraisal, err)
	}

	appraisal, err := o.initEvidenceContext(stHandler, token)
	if err != nil {
		return o.finalize(appraisal, err)
	}

	tas, err := o.getTrustAnchors(appraisal.EvidenceContext.TrustAnchorIds)
	if err != nil {
		if errors.Is(err, kvstore.ErrKeyNotFound) {
			err = handlermod.BadEvidence("no trust anchor for %s",
				appraisal.EvidenceContext.TrustAnchorIds)
			appraisal.SetAllClaims(ear.CryptoValidationFailedClaim)
			appraisal.AddPolicyClaim("problem", "no trust anchor for evidence")
		}
		return o.finalize(appraisal, err)
	}

	claims, err := handler.ExtractClaims(token, tas)
	if err != nil {
		if errors.Is(err, handlermod.BadEvidenceError{}) {
			appraisal.AddPolicyClaim("problem", err.Error())
		}
		return o.finalize(appraisal, err)
	}

	referenceIDs, err := stHandler.GetRefValueIDs(token.TenantId, tas, claims)
	if err != nil {
		return o.finalize(appraisal, err)
	}

	appraisal.EvidenceContext.Evidence, err = structpb.NewStruct(claims)
	if err != nil {
		err = fmt.Errorf("unserializable claims in result: %w", err)
		return o.finalize(appraisal, err)
	}

	appraisal.EvidenceContext.ReferenceIds = referenceIDs

	o.logger.Debugw("constructed evidence context",
		"software-id", appraisal.EvidenceContext.ReferenceIds,
		"trust-anchor-id", appraisal.EvidenceContext.TrustAnchorIds)

	var multEndorsements []string
	for _, refvalID := range appraisal.EvidenceContext.ReferenceIds {

		endorsements, err := o.EnStore.Get(refvalID)
		if err != nil && !errors.Is(err, kvstore.ErrKeyNotFound) {
			return o.finalize(appraisal, err)
		}

		o.logger.Debugw("obtained endorsements", "endorsements", endorsements)
		multEndorsements = append(multEndorsements, endorsements...)
	}

	if err = handler.ValidateEvidenceIntegrity(token, tas, multEndorsements); err != nil {
		if errors.Is(err, handlermod.BadEvidenceError{}) {
			appraisal.SetAllClaims(ear.CryptoValidationFailedClaim)
			appraisal.AddPolicyClaim("problem", "integrity validation failed")
		}
		return o.finalize(appraisal, err)
	}

	appraisedResult, err := handler.AppraiseEvidence(appraisal.EvidenceContext, multEndorsements)
	if err != nil {
		return o.finalize(appraisal, err)
	}
	appraisedResult.Nonce = appraisal.Result.Nonce
	appraisal.Result = appraisedResult
	appraisal.InitPolicyID()

	err = o.PolicyManager.Evaluate(ctx, handler.GetAttestationScheme(), appraisal, multEndorsements)
	if err != nil {
		return o.finalize(appraisal, err)
	}

	o.logger.Infow("evaluated attestation result", "attestation-result", appraisal.Result)

	return o.finalize(appraisal, nil)
}

func (c *GRPC) initEvidenceContext(
	handler handler.IStoreHandler,
	token *proto.AttestationToken,
) (*appraisal.Appraisal, error) {
	var err error

	appraisal := appraisal.New(token.TenantId, token.Nonce, handler.GetAttestationScheme())
	appraisal.EvidenceContext.TrustAnchorIds, err = handler.GetTrustAnchorIDs(token)

	if errors.Is(err, handlermod.BadEvidenceError{}) {
		appraisal.SetAllClaims(ear.CryptoValidationFailedClaim)
		appraisal.AddPolicyClaim("problem", "could not establish identity from evidence")
	}

	return appraisal, err
}

func (c *GRPC) getTrustAnchors(id []string) ([]string, error) {

	var taValues []string //nolint
	for _, taID := range id {
		values, err := c.TaStore.Get(taID)
		if err != nil {
			return []string{""}, err
		}

		// For now, Veraison schemes only support one trust anchor per trustAnchorID
		if len(values) != 1 {
			return []string{""}, fmt.Errorf("found %d trust anchors, want 1", len(values))
		}
		taValues = append(taValues, values[0])
	}

	return taValues, nil
}

func (c *GRPC) GetSupportedVerificationMediaTypes(context.Context, *emptypb.Empty) (*proto.MediaTypeList, error) {
	mts := c.EvPluginManager.GetRegisteredMediaTypes()
	return &proto.MediaTypeList{MediaTypes: mts}, nil
}

func (c *GRPC) GetSupportedProvisioningMediaTypes(context.Context, *emptypb.Empty) (*proto.MediaTypeList, error) {
	mts := c.EndPluginManager.GetRegisteredMediaTypes()
	return &proto.MediaTypeList{MediaTypes: mts}, nil
}

func (o *GRPC) GetEARSigningPublicKey(context.Context, *emptypb.Empty) (*proto.PublicKey, error) {
	alg, key, err := o.EarSigner.GetEARSigningPublicKey()
	if err != nil {
		return nil, err
	}

	err = key.Set("alg", alg.String())
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

func (o *GRPC) finalize(
	appraisal *appraisal.Appraisal,
	err error,
) (*proto.AppraisalContext, error) {
	var signErr error

	if err != nil {
		if errors.Is(err, handler.BadEvidenceError{}) {
			// NOTE(setrofim): I debated whether this should be
			// logged as Info or Warn. Ultimately deciding to go
			// with Warn, to make it easier to identifier the
			// failed validations in a stream of successful ones.
			// As we're effectively "swallowing" the error here,
			// and the response the client receives with only
			// contain "unexpected evidence", or whatever, and not
			// have the details, this log line can be important in
			// debugging problems.
			o.logger.Warn(err)
			// Clear the error as we've "handled" by setting the
			// claim in the result.
			err = nil
		} else {
			o.logger.Error(err)
			appraisal.SetAllClaims(ear.VerifierMalfunctionClaim)
		}

	}

	appraisal.Result.UpdateStatusFromTrustVector()

	appraisal.SignedEAR, signErr = o.EarSigner.Sign(*appraisal.Result)
	if signErr != nil {
		// Signing error overrides whatever the problem that got us
		// here was, as it indicates a serious issue with the service.
		// TODO(setrofim): signing an EAR should be an error-free
		// operation. It should either succeed or panic if there is a
		// serious issue with the underlying platform (such as OOM).
		// Any other problems (e.g. bad/missing key) should be
		// identified and handled during service initialisation.
		err = signErr
	}

	return appraisal.GetContext(), err
}
