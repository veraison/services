// Copyright 2022-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package trustedservices

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net"
	"os"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/veraison/corim/coserv"
	"github.com/veraison/ear"
	"github.com/veraison/services/config"
	"github.com/veraison/services/handler"
	handlermod "github.com/veraison/services/handler"
	"github.com/veraison/services/kvstore"
	"github.com/veraison/services/plugin"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/vts/appraisal"
	"github.com/veraison/services/vts/coservsigner"
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
type GRPCConfig struct {
	ServerAddress string   `mapstructure:"server-addr" valid:"dialstring"`
	ListenAddress string   `mapstructure:"listen-addr" valid:"dialstring" config:"zerodefault"`
	UseTLS        bool     `mapstructure:"tls" config:"zerodefault"`
	ServerCert    string   `mapstructure:"cert" config:"zerodefault"`
	ServerCertKey string   `mapstructure:"cert-key" config:"zerodefault"`
	CACerts       []string `mapstructure:"ca-certs" config:"zerodefault"`
	DispatchTable string   `mapstructure:"dispatch-table" config:"zerodefault"`
}

func NewGRPCConfig() *GRPCConfig {
	return &GRPCConfig{ServerAddress: DefaultVTSAddr}
}

type GRPC struct {
	ServerAddress string

	TaStore                   kvstore.IKVStore
	EnStore                   kvstore.IKVStore
	EvPluginManager           plugin.IManager[handler.IEvidenceHandler]
	EndPluginManager          plugin.IManager[handler.IEndorsementHandler]
	StorePluginManager        plugin.IManager[handler.IStoreHandler]
	CoservProxyPluginManager  plugin.IManager[handler.ICoservProxyHandler]
	LeadVerifierPluginManager plugin.IManager[handler.IComponentVerifierClientHandler]
	PolicyManager             *policymanager.PolicyManager
	EarSigner                 earsigner.IEarSigner
	CoservSigner              coservsigner.ICoservSigner
	CACertsPEM                [][]byte

	Server *grpc.Server
	Socket net.Listener

	logger *zap.SugaredLogger

	proto.UnimplementedVTSServer
}

func NewGRPC(
	taStore, enStore kvstore.IKVStore,
	evidencePluginManager plugin.IManager[handler.IEvidenceHandler],
	endorsementPluginManager plugin.IManager[handler.IEndorsementHandler],
	leadVerifierPluginManager plugin.IManager[handler.IComponentVerifierClientHandler],
	storePluginManager plugin.IManager[handler.IStoreHandler],
	coservProxyPluginManager plugin.IManager[handler.ICoservProxyHandler],
	policyManager *policymanager.PolicyManager,
	earSigner earsigner.IEarSigner,
	coservSigner coservsigner.ICoservSigner,
	logger *zap.SugaredLogger,
) ITrustedServices {
	return &GRPC{
		TaStore:                   taStore,
		EnStore:                   enStore,
		EvPluginManager:           evidencePluginManager,
		EndPluginManager:          endorsementPluginManager,
		LeadVerifierPluginManager: leadVerifierPluginManager,
		StorePluginManager:        storePluginManager,
		CoservProxyPluginManager:  coservProxyPluginManager,
		PolicyManager:             policyManager,
		EarSigner:                 earSigner,
		CoservSigner:              coservSigner,
		logger:                    logger,
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
	evidenceManager plugin.IManager[handler.IEvidenceHandler],
	endorsementManager plugin.IManager[handler.IEndorsementHandler],
	storeManager plugin.IManager[handler.IStoreHandler],
	coservProxyManager plugin.IManager[handler.ICoservProxyHandler],
) error {
	var err error

	cfg := GRPCConfig{
		ServerAddress: DefaultVTSAddr,
		UseTLS:        true,
	}

	loader := config.NewLoader(&cfg)
	if err := loader.LoadFromViper(v); err != nil {
		return err
	}

	o.EvPluginManager = evidenceManager
	o.EndPluginManager = endorsementManager
	o.StorePluginManager = storeManager
	o.CoservProxyPluginManager = coservProxyManager

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

	var opts []grpc.ServerOption

	var certsPEM [][]byte
	if len(cfg.CACerts) > 0 {
		var err error
		certsPEM, err = LoadCACerts(cfg.CACerts)
		if err != nil {
			return err
		}
	} else {
		certsPEM = [][]byte{}
	}
	o.CACertsPEM = certsPEM

	if cfg.UseTLS {
		o.logger.Info("loading TLS credentials")
		creds, err := LoadTLSCreds(cfg.ServerCert, cfg.ServerCertKey, cfg.CACerts)
		if err != nil {
			return err
		}

		opts = append(opts, grpc.Creds(creds))
	}

	if cfg.DispatchTable != "" {
		if err := LoadDispatchTable(cfg.DispatchTable); err != nil {
			return err
		}
	}

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

	if err := o.CoservProxyPluginManager.Close(); err != nil {
		o.logger.Errorf("coserv plugin manager shutdown failed: %v", err)
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

	if o.CoservSigner != nil {
		if err := o.CoservSigner.Close(); err != nil {
			o.logger.Errorf("CoSERV signer closure failed: %v", err)
		}
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

	// Serialize the CA cert pool for transmission if it's a signed CoRIM
	var caCertPoolBytes []byte
	mt, _, err := mime.ParseMediaType(req.MediaType)
	if err != nil {
		return nil, err
	}
	if mt == "application/rim+cose" && len(o.CACertsPEM) != 0 {
		var err error
		caCertPoolBytes, err = SerializeCertPEMBytes(o.CACertsPEM)
		if err != nil {
			return nil, err
		}
	}

	rsp, err := handlerPlugin.Decode(req.Data, req.MediaType, caCertPoolBytes)
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

func (o *GRPC) GetCompositeAttestation(
	ctx context.Context,
	token *proto.AttestationToken,
) (*proto.AppraisalContext, error) {
	o.logger.Infow("get composite attestation", "media-type", token.MediaType,
		"tenant-id", token.TenantId)

	// TODO(tho)
	//
	//	lead_verifier(CE, ce-type, n) -> CAR {
	//		// lookup CE parser
	//		parser = ce_parsers_table[ce-type]
	//
	//		// tokenise the composite evidence
	//		{CE_i, ce_i-type, label_i} = parser(CE)
	//
	//		// walk the items in the composite evidence
	//		foreach c, t, l in {CE_i, ce_i-type, label_i}:
	//			client = dispatch_table[ce_i-type]
	//			if !client:
	//				EAR[label_i] = { raw-evidence: c, status: unknown }
	//			else:
	//				EAR[label_i] = client(c, n)
	//
	//		CAR = make_car(EAR, lead_verifier_signing_key)
	//
	//		return CAR
	//	}

	return nil, errors.New("not implemented")
}

func (o *GRPC) GetSupportedCompositeEvidenceMediaTypes(context.Context, *emptypb.Empty) (*proto.MediaTypeList, error) {

	// Look for all parsers
	// Invoke each Parsers to get the Collection MediaTypes
	// return it inside MediaTypeList
	return nil, nil
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
			var badErr handlermod.BadEvidenceError
			claimStr := "integrity validation failed"
			ok := errors.As(err, &badErr)
			if ok {
				claimStr += fmt.Sprintf(": %s", badErr.ToString())
			}
			appraisal.SetAllClaims(ear.CryptoValidationFailedClaim)
			appraisal.AddPolicyClaim("problem", claimStr)
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

func (c *GRPC) assembleCoservMediaTypes(mts []string, filter string) []string {
	mediaTypes := make([]string, 0, 10)

	for _, mt := range mts {
		t, p, err := mime.ParseMediaType(mt)
		if err != nil || t != filter {
			continue
		}
		profile := p["profile"]
		if profile == "" {
			continue
		}

		mediaType := fmt.Sprintf(`application/coserv+cbor; profile=%q`, profile)
		mediaTypes = append(mediaTypes, mediaType)

		if c.CoservSigner != nil {
			mediaType = fmt.Sprintf(`application/coserv+cose; profile=%q`, profile)
			mediaTypes = append(mediaTypes, mediaType)
		}
	}

	return mediaTypes
}

func (c *GRPC) GetSupportedCoservMediaTypes(context.Context, *emptypb.Empty) (*proto.MediaTypeList, error) {
	var mediaTypes []string

	corimDerived := c.assembleCoservMediaTypes(
		c.EndPluginManager.GetRegisteredMediaTypes(),
		"application/rim+cbor",
	)

	mediaTypes = append(mediaTypes, corimDerived...)

	coservProxyDerived := c.assembleCoservMediaTypes(
		c.CoservProxyPluginManager.GetRegisteredMediaTypes(),
		"application/coserv+cbor",
	)

	mediaTypes = append(mediaTypes, coservProxyDerived...)

	c.logger.Debugw("GetSupportedCoservMediaTypes", "media types", mediaTypes)

	return &proto.MediaTypeList{MediaTypes: mediaTypes}, nil
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

func (o *GRPC) GetCoservSigningPublicKey(context.Context, *emptypb.Empty) (*proto.PublicKey, error) {
	// If CoSERV is not enabled, return an empty key.
	if o.CoservSigner == nil {
		return &proto.PublicKey{Key: ""}, nil
	}

	alg, key, err := o.CoservSigner.GetCoservSigningPublicKey()
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

func getEndorsementsError(err error) *proto.EndorsementQueryOut {
	return &proto.EndorsementQueryOut{
		Status: &proto.Status{
			Result:      false,
			ErrorDetail: fmt.Sprintf("%v", err),
		},
	}
}

func (o *GRPC) getEndorsementsFromStores(query *proto.EndorsementQueryIn) ([]byte, error) {
	var q coserv.Coserv
	if err := q.FromBase64Url(query.Query); err != nil {
		return nil, err
	}

	// select store based on the requested artefact type
	var (
		store     kvstore.IKVStore
		storeName string
	)
	switch q.Query.ArtifactType {
	case coserv.ArtifactTypeEndorsedValues:
		return nil, errors.New("endorsed value queries are not supported")
	case coserv.ArtifactTypeReferenceValues:
		storeName = "reference-value"
		store = o.EnStore
	case coserv.ArtifactTypeTrustAnchors:
		storeName = "trust-anchors"
		store = o.TaStore
	}

	profile, err := q.Profile.Get()
	if err != nil {
		return nil, err
	}

	// Look up a matching endorsement plugin
	endorsementHandler, err := o.EndPluginManager.LookupByMediaType(fmt.Sprintf(`application/rim+cbor; profile=%q`, profile))
	if err != nil {
		return nil, err
	}

	scheme := endorsementHandler.GetAttestationScheme()

	storeHandler, err := o.StorePluginManager.LookupByAttestationScheme(scheme)
	if err != nil {
		return nil, err
	}

	queryKeys, err := storeHandler.SynthCoservQueryKeys(DummyTenantID, query.Query)
	if err != nil {
		return nil, err
	}

	var resultSet []string
	for _, key := range queryKeys {
		res, err := store.Get(key)
		if err != nil && !errors.Is(err, kvstore.ErrKeyNotFound) {
			return nil, fmt.Errorf("lookup %q in %s failed: %w", key, storeName, err)
		}

		resultSet = append(resultSet, res...)
	}

	coservResults, err := endorsementHandler.CoservRepackage(query.Query, resultSet)
	if err != nil {
		return nil, err
	}

	return coservResults, nil
}

func (o *GRPC) getEndorsementsFromProxy(handlerPlugin handler.ICoservProxyHandler, query *proto.EndorsementQueryIn) ([]byte, error) {
	return handlerPlugin.GetEndorsements(DummyTenantID, query.Query)
}

func (o *GRPC) GetEndorsements(ctx context.Context, query *proto.EndorsementQueryIn) (*proto.EndorsementQueryOut, error) {
	o.logger.Debugw("GetEndorsements", "media-type", query.MediaType)

	var (
		err           error
		out           []byte
		handlerPlugin handler.ICoservProxyHandler
	)

	// First, check to see if we have a CoSERV proxy plugin that can handle this query
	handlerPlugin, err = o.CoservProxyPluginManager.LookupByMediaType(query.MediaType)
	if err == nil {
		// No error means we have a proxy plugin, so delegate to that.
		out, err = o.getEndorsementsFromProxy(handlerPlugin, query)
	} else {
		// There was no proxy plugin, so assume we can obtain from own stores
		out, err = o.getEndorsementsFromStores(query)
	}

	if err != nil {
		return getEndorsementsError(err), nil
	}

	// If we have a CoSERV signer configured and the client requested a COSE
	// response, sign the result here.
	if strings.HasPrefix(query.MediaType, "application/coserv+cose") {
		if o.CoservSigner != nil {
			var tmp coserv.Coserv
			err = tmp.FromCBOR(out)
			if err != nil {
				return getEndorsementsError(fmt.Errorf("could not parse CoSERV response: %w", err)), nil
			}

			out, err = o.CoservSigner.Sign(tmp)
			if err != nil {
				return getEndorsementsError(fmt.Errorf("could not sign CoSERV response: %w", err)), nil
			}
		} else {
			return getEndorsementsError(errors.New("no CoSERV signer configured")), nil
		}
	}

	return &proto.EndorsementQueryOut{
		Status:    &proto.Status{Result: true},
		ResultSet: out,
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

func LoadTLSCreds(
	certPath, keyPath string,
	caPaths []string,
) (credentials.TransportCredentials, error) {
	if certPath == "" {
		return nil, fmt.Errorf("cert path must be specified when TLS is enabled")
	}

	if keyPath == "" {
		return nil, fmt.Errorf("cert key path must be specified when TLS is enabled")
	}

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("error loading cert key pair: %w", err)
	}

	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("error loading system certs: %w", err)
	}

	for _, caPath := range caPaths {
		caCertPEM, err := os.ReadFile(caPath)
		if err != nil {
			return nil, fmt.Errorf("error reading CA cert in %s: %w", caPath, err)
		}

		if !certPool.AppendCertsFromPEM(caCertPEM) {
			return nil, fmt.Errorf("invalid CA cert in %s", caPath)
		}
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		RootCAs:      certPool,
		ClientCAs:    certPool,
		MinVersion:   tls.VersionTLS12,
	}

	return credentials.NewTLS(config), nil
}

// LoadCaCerts loads and validates CA certificates from file paths
func LoadCACerts(paths []string) ([][]byte, error) {
	certsPEM := [][]byte{}

	for _, path := range paths {
		certPEM, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("error reading cert in %s: %w", path, err)
		}

		tempPool := x509.NewCertPool()
		if !tempPool.AppendCertsFromPEM(certPEM) {
			return nil, fmt.Errorf("invalid cert in %s", path)
		}

		certsPEM = append(certsPEM, certPEM)
	}

	return certsPEM, nil
}

// SerializeCertPEMBytes converts the CA certificate pool to PEM format for transmission
func SerializeCertPEMBytes(certPEMs [][]byte) ([]byte, error) {
	if len(certPEMs) == 0 {
		return []byte{}, nil
	}

	var allPEM bytes.Buffer
	for _, pemData := range certPEMs {
		if _, err := allPEM.Write(pemData); err != nil {
			return nil, fmt.Errorf("failed to write certificate data: %w", err)
		}
	}

	return allPEM.Bytes(), nil
}
