// Copyright 2022-2026 Contributors to the Veraison project.
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

	"github.com/veraison/corim-store/pkg/model"
	corimstore "github.com/veraison/corim-store/pkg/store"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/corim"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/ear"
	"github.com/veraison/services/config"
	handlermod "github.com/veraison/services/handler"
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


var ErrMeasurementsNotSupported = errors.New("measurements in CoSERV queries are not supported")

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
}

func NewGRPCConfig() *GRPCConfig {
	return &GRPCConfig{ServerAddress: DefaultVTSAddr}
}

type GRPC struct {
	ServerAddress string

	Store			 *corimstore.Store
	SchemePluginManager      plugin.IManager[handlermod.ISchemeHandler]
	CoservProxyPluginManager plugin.IManager[handlermod.ICoservProxyHandler]
	PolicyManager            *policymanager.PolicyManager
	EarSigner                earsigner.IEarSigner
	CoservSigner             coservsigner.ICoservSigner
	rootCerts                *x509.CertPool

	Server *grpc.Server
	Socket net.Listener

	logger *zap.SugaredLogger

	proto.UnimplementedVTSServer
}

func NewGRPC(
	store *corimstore.Store,
	schemePluginManager plugin.IManager[handlermod.ISchemeHandler],
	coservProxyPluginManager plugin.IManager[handlermod.ICoservProxyHandler],
	policyManager *policymanager.PolicyManager,
	earSigner earsigner.IEarSigner,
	coservSigner coservsigner.ICoservSigner,
	logger *zap.SugaredLogger,
) ITrustedServices {
	return &GRPC{
		Store: 			  store,
		SchemePluginManager:      schemePluginManager,
		CoservProxyPluginManager: coservProxyPluginManager,
		PolicyManager:            policyManager,
		EarSigner:                earSigner,
		CoservSigner:             coservSigner,
		logger:                   logger,
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

	o.logger.Info("loading root CA certs")
	o.rootCerts, err = LoadCACerts(cfg.CACerts)
	if err != nil {
		return err
	}

	if cfg.UseTLS {
		o.logger.Info("loading TLS credentials")
		creds, err := LoadTLSCreds(cfg.ServerCert, cfg.ServerCertKey, cfg.CACerts)
		if err != nil {
			return err
		}

		opts = append(opts, grpc.Creds(creds))
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

	if err := o.SchemePluginManager.Close(); err != nil {
		o.logger.Errorf("scheme plugin manager shutdown failed: %v", err)
	}

	if err := o.CoservProxyPluginManager.Close(); err != nil {
		o.logger.Errorf("coserv plugin manager shutdown failed: %v", err)
	}

	if err := o.Store.Close(); err != nil {
		o.logger.Errorf("store closure failed: %v", err)
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
	mediaTypes := o.SchemePluginManager.GetRegisteredMediaTypes()

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

func (o *GRPC) SubmitEndorsements(
	ctx context.Context,
	req *proto.SubmitEndorsementsRequest,
) (*proto.SubmitEndorsementsResponse, error) {
	o.logger.Debugw("SubmitEndorsements", "media-type", req.MediaType)

	mt, mtParams, err := mime.ParseMediaType(req.MediaType)
	if err != nil {
		return nil, err
	}
	profile := mtParams["profile"]

	var uc *corim.UnsignedCorim
	switch mt {
	case "application/rim+cose":
		uc, err = o.decodeAndValidateSignedCorim(req.Data)
		if err != nil {
			return submitEndorsementErrorResponse(err), nil
		}
	case "application/rim+cbor":
		uc, err = o.decodeAndValidateUnsignedCorim(req.Data)
		if err != nil {
			return submitEndorsementErrorResponse(err), nil
		}
	default:
		err = fmt.Errorf("unsupported media type: %s", req.MediaType)
		return submitEndorsementErrorResponse(err), nil
	}

	if uc.Profile == nil {
		return nil, errors.New("profile not set in CoRIM")
	}

	ucProfile, err :=  uc.Profile.Get()
	if err != nil {
		return nil, fmt.Errorf("invalid profile in CoRIM: %v", uc.Profile)
	}

	o.logger.Debugw("    CoRIM profile", "profile", ucProfile)

	if ucProfile != profile {
		err := fmt.Errorf(
			"CoRIM profile (%s) does not match media type profile (%s)",
			ucProfile,
			profile,
		)
		o.logger.Warn(err.Error())
		return nil, err
	}

	handlerPlugin, err := o.SchemePluginManager.LookupByMediaType(req.MediaType)
	if err != nil {
		return nil, err
	}

	resp, err := handlerPlugin.ValidateCorim(uc)
	if err != nil {
		return nil, err
	} else if ! resp.IsValid {
		return submitEndorsementErrorResponse(resp.Error()), nil
	}

	label := fmt.Sprintf("%s/%s", DummyTenantID, handlerPlugin.GetAttestationScheme())
	digest := o.Store.Digest(req.Data)
	if err := o.storeEndorsements(ctx, uc, label, digest); err != nil {
		return submitEndorsementErrorResponse(err), nil
	}

	return submitEndorsementSuccessResponse(), nil
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

func (o *GRPC) decodeAndValidateSignedCorim(data []byte) (*corim.UnsignedCorim, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty corim data")
	}

	// Parse the signed CoRIM which extracts certificate chain automatically through extractX5Chain
	sc, err := corim.UnmarshalAndValidateSignedCorimFromCBOR(data)
	if  err != nil {
		return nil, fmt.Errorf("failed to parse signed CoRIM: %w", err)
	}

	if sc.SigningCert == nil {
		return nil, fmt.Errorf("no signing certificate found in the CoRIM")
	}

	intermediateCertPool := x509.NewCertPool()
	for _, cert := range sc.IntermediateCerts {
		intermediateCertPool.AddCert(cert)
	}

	// Verify the certificate chain with properly separated root and intermediate pools
	verifyOpts := x509.VerifyOptions{
		Roots:         o.rootCerts,
		Intermediates: intermediateCertPool,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}

	_, err = sc.SigningCert.Verify(verifyOpts)
	if err != nil {
		return nil, fmt.Errorf("certificate chain verification failed: %w", err)
	}

	// Verify the signature using the signing certificate's public key
	if err := sc.Verify(sc.SigningCert.PublicKey); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	return &sc.UnsignedCorim, nil
}

func (o *GRPC) decodeAndValidateUnsignedCorim(data []byte) (*corim.UnsignedCorim, error) {
	return corim.UnmarshalAndValidateUnsignedCorimFromCBOR(data)
}

func (o *GRPC) storeEndorsements(
	_ context.Context,
	uc *corim.UnsignedCorim,
	label string,
	digest []byte,
) error {
	manifest, err := model.NewManifestFromCoRIM(uc)
	if err != nil {
		return err
	}
	manifest.Label = label
	manifest.Digest = digest
	manifest.SetActive(true)

	if err := o.Store.AddManifest(manifest); err != nil {
		return err
	}

	return nil
}

func (o *GRPC) GetAttestation(
	ctx context.Context,
	token *proto.AttestationToken,
) (*proto.AppraisalContext, error) {
	evidence := appraisal.NewEvidenceFromProtobuf(token)
	o.logger.Infow("get attestation", "media-type", evidence.MediaType,
		"tenant-id", evidence.TenantID)

	appraisal := appraisal.NewContext(evidence)

	handler, err := o.SchemePluginManager.LookupByMediaType(evidence.MediaType)
	if err != nil {
		appraisal.SetAllClaims(ear.UnexpectedEvidenceClaim)
		appraisal.AddPolicyClaim("problem", "could not resolve media type")
		return o.finalize(appraisal, err)
	}

	if err := appraisal.SetScheme(handler.GetAttestationScheme()); err != nil {
		return o.finalize(appraisal, err)
	}

	appraisal.TrustAnchorIDs, err = handler.GetTrustAnchorIDs(evidence)
	if err != nil {
		if errors.Is(err, handlermod.BadEvidenceError{}) {
			appraisal.SetAllClaims(ear.CryptoValidationFailedClaim)
			appraisal.AddPolicyClaim("problem", "could not establish identity from evidence")
		}

		return o.finalize(appraisal, err)
	}

	// TODO(setrofim): in principle, we should be matching exactly
	// here, as we are trying to identify triples for a specific
	// entity described by the taID environment. However, CoRIM does
	// not provide a good way to endorse an environment description. So
	// in paractice, enviroments inside CoRIMs may serve
	// the double duty of providing a way to match measurements
	// _and_ an additional decription of the enviroments that is
	// not present in the evidence.
	//
	// Specifically, our parsec-tpm scheme currently relies on the
	// fact that the trust anchors are provisioned with enviroment
	// that contains a class-id as well as an instance-id; the
	// evidence only contains an instance-id, and the class-id in
	// the trust anchor is then used to retrieve reference values;
	// there is no way to directly link reference values to the
	// evidence.
	//
	// Because provisioned enviroments may containd "additional"
	// descriptive elements, as well as elements used for matching,
	// we are forced to do inexact matching here for now, and leave
	// it to the attestation schemes to resolve this.
	matchExactly := false
	trustAnchors, err := o.getKeyTriples(appraisal.StoreLabel(), appraisal.TrustAnchorIDs, matchExactly)
	if err != nil {
		if errors.Is(err, corimstore.ErrNoMatch) {
			err = handlermod.BadEvidence("no trust anchor for %s", appraisal.DescribeTrustAnchorIDs())
			appraisal.SetAllClaims(ear.CryptoValidationFailedClaim)
			appraisal.AddPolicyClaim("problem", "no trust anchor for evidence")
		}

		return o.finalize(appraisal, err)
	}

	claims, err := handler.ExtractClaims(appraisal.Evidence, trustAnchors)
	if err != nil {
		if errors.Is(err, handlermod.BadEvidenceError{}) {
			appraisal.AddPolicyClaim("problem", err.Error())
		}
		return o.finalize(appraisal, err)
	}
	appraisal.Claims = claims

	appraisal.ReferenceValueIDs, err = handler.GetReferenceValueIDs(trustAnchors, claims)
	if err != nil {
		return o.finalize(appraisal, err)
	}

	o.logger.Debugw("constructed evidence context",
		"software-id", appraisal.ReferenceValueIDs,
		"trust-anchor-id", appraisal.TrustAnchorIDs)

	o.logger.Debug("obtaining endrosements...")
	endorsements, err := o.getValueTriples(appraisal.StoreLabel(), appraisal.ReferenceValueIDs, true)
	if err != nil {
		return o.finalize(appraisal, err)
	}

	o.logger.Debug("validating evidence...")
	if err = handler.ValidateEvidenceIntegrity(appraisal.Evidence, trustAnchors, endorsements); err != nil {
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

	o.logger.Debug("appraising claims...")
	appraisedResult, err := handler.AppraiseClaims(claims, endorsements)
	if err != nil {
		return o.finalize(appraisal, err)
	}
	appraisedResult.Nonce = appraisal.Result.Nonce
	appraisal.Result = appraisedResult
	appraisal.InitPolicyID()

	o.logger.Debug("evaluating policy...")
	err = o.PolicyManager.Evaluate(ctx, appraisal, endorsements)
	if err != nil {
		return o.finalize(appraisal, err)
	}

	o.logger.Infow("evaluated attestation result", "attestation-result", appraisal.Result)
	return o.finalize(appraisal, nil)
}

func (o *GRPC) getKeyTriples(
	label string,
	trustAnchorIDs []*comid.Environment,
	exact bool,
) ([]*comid.KeyTriple, error) {
	var keyTriples []*comid.KeyTriple //nolint

	for _, taID := range trustAnchorIDs {
		env, err := model.NewEnvironmentFromCoRIM(taID)
		if err != nil {
			return nil, err
		}

		triples, err := o.Store.GetActiveKeyTriples(env, label, exact)
		if err != nil {
			return nil, err
		}

		for _, triple := range triples {
			compiled, err := triple.ToCoRIM()
			if err != nil {
				return nil, err
			}

			keyTriples = append(keyTriples, compiled)
		}
	}

	return keyTriples, nil
}

func (o *GRPC) getValueTriples(
	label string,
	referenceValueIDs []*comid.Environment,
	exact bool,
) ([]*comid.ValueTriple, error) {
	var valueTriples []*comid.ValueTriple //nolint

	for _, valID := range referenceValueIDs {
		env, err := model.NewEnvironmentFromCoRIM(valID)
		if err != nil {
			return nil, err
		}

		triples, err := o.Store.GetActiveValueTriples(env, label, exact)
		if err != nil && !errors.Is(err, corimstore.ErrNoMatch){
			return nil, err
		}

		for _, triple := range triples {
			compiled, err := triple.ToCoRIM()
			if err != nil {
				return nil, err
			}

			valueTriples = append(valueTriples, compiled)
		}
	}

	return valueTriples, nil
}

func (c *GRPC) GetSupportedVerificationMediaTypes(context.Context, *emptypb.Empty) (*proto.MediaTypeList, error) {
	mts := c.SchemePluginManager.GetRegisteredMediaTypesByCategory("verification")
	return &proto.MediaTypeList{MediaTypes: mts}, nil
}

func (c *GRPC) GetSupportedProvisioningMediaTypes(context.Context, *emptypb.Empty) (*proto.MediaTypeList, error) {
	mts := c.SchemePluginManager.GetRegisteredMediaTypesByCategory("provisioning")
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
		c.SchemePluginManager.GetRegisteredMediaTypes(), // TODO! (endorsement only)
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

func (o *GRPC) getEndorsementsFromStores(queryIn *proto.EndorsementQueryIn) ([]byte, error) {
	var query coserv.Coserv
	if err := query.FromBase64Url(queryIn.Query); err != nil {
		return nil, err
	}

	profile, err := query.Profile.Get()
	if err != nil {
		return nil, err
	}

	// Look up a matching endorsement plugin
	schemeHandler, err := o.SchemePluginManager.LookupByMediaType(
			fmt.Sprintf(`application/rim+cbor; profile=%q`, profile))
	if err != nil {
		return nil, err
	}

	scheme := schemeHandler.GetAttestationScheme()
	label := fmt.Sprintf("%s/%s", DummyTenantID, scheme)

	authority, err := comid.NewCryptoKeyTaggedBytes([]byte("dummyauth"))
	if err != nil {
		return nil, err
	}

	environments, err := querySelectorToEnvironments(&query.Query.EnvironmentSelector)
	if err != nil {
		return nil, err
	}

	resultSet := coserv.NewResultSet()
	switch query.Query.ArtifactType {
	case coserv.ArtifactTypeTrustAnchors:
		keyTriples, err := o.getKeyTriples(label, environments, false)
		if err != nil {
			return nil, err
		}

		for _,  keyTriple := range keyTriples {
			resultSet.AddAttestationKeys(coserv.AKQuad{
				Authorities: comid.NewCryptoKeys().Add(authority),
				AKTriple: keyTriple,
			})
		}
	case coserv.ArtifactTypeReferenceValues:
		valueTriples, err := o.getValueTriples(label, environments, false)
		if err != nil {
			return nil, err
		}

		for _,  valueTriple := range valueTriples {
			resultSet.AddReferenceValues(coserv.RefValQuad{
				Authorities: comid.NewCryptoKeys().Add(authority),
				RVTriple: valueTriple,
			})
		}
	default:
		return nil, errors.New("only reference value and trust anchors are supported at present")
	}

	if err := query.AddResults(*resultSet); err != nil {
		return nil, fmt.Errorf("could not add result set to query: %w", err)
	}

	return query.ToCBOR()
}

func (o *GRPC) getEndorsementsFromProxy(
	handlerPlugin handlermod.ICoservProxyHandler,
	query *proto.EndorsementQueryIn,
) ([]byte, error) {
	return handlerPlugin.GetEndorsements(DummyTenantID, query.Query)
}

func (o *GRPC) GetEndorsements(
	ctx context.Context,
	query *proto.EndorsementQueryIn,
) (*proto.EndorsementQueryOut, error) {
	o.logger.Debugw("GetEndorsements", "media-type", query.MediaType)

	var (
		err           error
		out           []byte
		handlerPlugin handlermod.ICoservProxyHandler
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

// finalize prepares the final appraisal context to be returned to the client.
// The EAR is signed using the verifier private key.
//
// The error parameter indicates whether there was an error during the
// attestation process. If a non-nil error is supplied, it is classified as a
// verifier malfunction - unless it's of type "bad evidence", in which case it
// is logged and the error is cleared because we assume the relevant claim has
// been already set in the attestation result.
func (o *GRPC) finalize(
	appraisal *appraisal.Context,
	err error,
) (*proto.AppraisalContext, error) {
	var signErr error

	if err != nil {
		if errors.Is(err, handlermod.BadEvidenceError{}) {
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

	pbAppraisal, pbErr := appraisal.ToProtobuf()
	if pbErr != nil {
		err = pbErr
	}

	return pbAppraisal, err
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

// LoadCaCerts loads and validates CA certificates from file paths, as well as the system certs.
func LoadCACerts(paths []string) (*x509.CertPool, error) {
	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("could not load system certs: %w", err)
	}

	if len(paths) == 0 {
		return certPool, nil
	}

	for _, path := range paths {
		certPEM, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("error reading cert in %s: %w", path, err)
		}

		if !certPool.AppendCertsFromPEM(certPEM) {
			return nil, fmt.Errorf("invalid cert in %s", path)
		}
	}

	return certPool, nil
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

func querySelectorToEnvironments(selector *coserv.EnvironmentSelector) ([]*comid.Environment, error) {
	var ret []*comid.Environment

	if selector.Classes != nil {
		for _, statefulClass := range *selector.Classes {
			if statefulClass.Measurements != nil {
				return nil, ErrMeasurementsNotSupported
			}

			ret = append(ret, &comid.Environment{
				Class: statefulClass.Class,
			})
		}
	}

	if selector.Instances != nil {
		for _, statefulInstance := range *selector.Instances {
			if statefulInstance.Measurements != nil {
				return nil, ErrMeasurementsNotSupported
			}

			ret = append(ret, &comid.Environment{
				Instance: statefulInstance.Instance,
			})
		}
	}

	if selector.Groups != nil {
		for _, statefulGroup := range *selector.Groups {
			if statefulGroup.Measurements != nil {
				return nil, ErrMeasurementsNotSupported
			}

			ret = append(ret, &comid.Environment{
				Group: statefulGroup.Group,
			})
		}
	}

	return ret, nil
}
