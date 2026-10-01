// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package ccacoservbroker

import (
	"bytes"
	"crypto"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/fxamacker/cbor/v2"
	jose "github.com/go-jose/go-jose/v4"
	coservBuilder "github.com/veraison/apiclient/coserv"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/coserv"
	cose "github.com/veraison/go-cose"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/plugin"
	"go.uber.org/zap"
)

const (
	pluginName = "cca-coserv-broker"
	schemeName = "ARM_CCA"

	//TODO: Update with newer cca platform profile when the corim dependency is updated
	ccaPlatformProfile         = "tag:arm.com,2025:cca_platform#1.0.0"
	signedCCAPlatformProfile   = `application/coserv+cose; profile="tag:arm.com,2025:cca_platform#1.0.0"`
	unSignedCCAPlatformProfile = `application/coserv+cbor; profile="tag:arm.com,2025:cca_platform#1.0.0"`
)

var (
	CoservMediaTypes = []string{
		`application/coserv+cbor; profile="tag:arm.com,2025:cca_platform#1.0.0"`,
	}
	SupportedCoservProfiles = []string{ccaPlatformProfile}
)

type StoreInitError struct {
	Reason string
}

func (e *StoreInitError) Error() string {
	return fmt.Sprintf("unable to load 'cca-coserv-broker' plugin due to error: %s", e.Reason)
}

func NewStoreInitError(reason string) error {
	return &StoreInitError{
		Reason: reason,
	}
}

func NewStoreInitErrorWithErr(reason string, err error) error {
	r := fmt.Sprintf(reason, err.Error())
	return &StoreInitError{
		Reason: r,
	}
}

func ErrFetchFailed(reason error) error {
	return fmt.Errorf("cca-coserv-broker store operation failed : with reason: %w", reason)
}

func ErrFetchFailedStr(reason string) error {
	return fmt.Errorf("cca-coserv-broker store operation failed : with reason: %s", reason)
}

type Broker struct {
	logger       *zap.SugaredLogger
	coservclient *coservBuilder.QueryConfig
	keys         []crypto.PublicKey
}

func NewBroker() *Broker {
	return &Broker{
		logger: log.Named(pluginName),
	}
}

func (b *Broker) Init(params *plugin.Parameters) error {
	if params != nil {
		enableBroker, ok := params.DefaultGet("enable-broker", false).(bool)
		if !ok {
			return NewStoreInitError("enable-broker config is not a string!")
		}
		if enableBroker {
			coservDiscovery := &coservBuilder.DiscoveryConfig{}
			coservclient := &coservBuilder.QueryConfig{}

			if discoveryURL, err := params.Get("coserv-url"); err != nil {
				return NewStoreInitError("coserv-url config is not provided!")
			} else {
				discoveryURL, ok := discoveryURL.(string)
				if !ok {
					return NewStoreInitError("coserv-url config is not a string!")
				}
				if err := coservDiscovery.SetDiscoveryURI(discoveryURL); err != nil {
					return NewStoreInitErrorWithErr("failed to set coserv-url in discovery client! : %s", err)
				}
			}

			if insecure, ok := params.DefaultGet("client-tls-insecure", false).(bool); ok {
				coservDiscovery.SetIsInsecure(insecure)
				coservclient.SetIsInsecure(insecure)
			} else {
				return NewStoreInitError("client-tls-insecure config is not a bool!")
			}

			var cacerts []string

			if certs, ok := params.DefaultGet("client-certs", "").(string); ok {
				cacerts = strings.FieldsFunc(
					certs,
					func(r rune) bool { return r == ',' },
				)
			} else {
				return NewStoreInitError("client-certs config is not a string!")
			}

			if len(cacerts) > 0 {
				if err := coservDiscovery.SetCerts(cacerts); err != nil {
					return NewStoreInitErrorWithErr("failed to add certs in discovery client! : %s", err)
				}
				if err := coservclient.SetCerts(cacerts); err != nil {
					return NewStoreInitErrorWithErr("failed to add certs in coserv client! : %s", err)
				}
			}

			if enablecache, ok := params.DefaultGet("client-caching", true).(bool); ok {
				coservDiscovery.EnableCache(enablecache)
				coservclient.EnableCache(enablecache)
			} else {
				return NewStoreInitError("client-caching config is not a bool!")
			}

			discoveryResult, err := coservDiscovery.Run()
			if err != nil {
				return NewStoreInitErrorWithErr("failed to fetch discovery document! : %s", err)
			}

			b.logger.Info("Successfuly fetched Discovery Document")

			if err := coservclient.SetRequestResponseURI(discoveryResult.QueryEndpointURL); err != nil {
				return NewStoreInitErrorWithErr("failed to set coserv query endpoint url in coserv client! : %s", err)
			}

			if unsignedcoserv, ok := params.DefaultGet("unsigned-coserv", false).(bool); ok {
				inner := func(mt string) error {
					hasCapability := false
					for _, capability := range discoveryResult.CapabilitiesList {
						if capability.MediaType == mt {
							hasCapability = true
							break
						}
					}
					if !hasCapability {
						return NewStoreInitError(fmt.Sprintf("remote coserv service is unable to serve %s mediatype response", mt))
					}
					return nil
				}
				if unsignedcoserv {
					if err := inner(unSignedCCAPlatformProfile); err != nil {
						return err
					}
				} else {
					if err := inner(signedCCAPlatformProfile); err != nil {
						return err
					}

					switch {
					case len(discoveryResult.VerificationKeyJwk) > 0:
						for _, key := range discoveryResult.VerificationKeyJwk {
							var jwk jose.JSONWebKey
							if err := jwk.UnmarshalJSON(key); err != nil {
								return NewStoreInitErrorWithErr("failed to parse JWK key in discovery document! : %s", err)
							}
							if pb, ok := jwk.Key.(crypto.PublicKey); ok {
								b.keys = append(b.keys, pb)
							} else {
								return NewStoreInitError("unable to convert JWK key to crypto PublicKey format!")
							}
						}
					case len(discoveryResult.VerificationKeyCose) > 0:
						for _, key := range discoveryResult.VerificationKeyCose {
							var coseKey cose.Key
							if err := coseKey.UnmarshalCBOR(key); err != nil {
								return NewStoreInitErrorWithErr("failed to parse COSE key in discovery document! : %s", err)
							}
							if pb, err := coseKey.PublicKey(); err == nil {
								b.keys = append(b.keys, pb)
							} else {
								return NewStoreInitErrorWithErr("unable to convert COSE key to crypto PublicKey format! : %s", err)
							}
						}
					default:
						return NewStoreInitError("signed coserv is configured for result but no keys found in discovery document!")
					}
				}
			} else {
				return NewStoreInitError("unsigned-coserv config is not bool!")
			}

			b.coservclient = coservclient

		} else {
			return NewStoreInitError("cca-coserv-broker is disabled!")
		}
	} else {
		return NewStoreInitError("parameters should be non-nil")
	}
	return nil
}

func (b *Broker) GetName() string {
	return pluginName
}

func (b *Broker) GetAttestationScheme() string {
	return schemeName
}

func (b *Broker) GetSupportedMediaTypes() map[string][]string {
	return map[string][]string{
		"coserv": CoservMediaTypes,
	}
}

// Fetches and validates CoSERV response
func fetchCoservResponse(broker *Broker, query *coserv.Coserv) (*coserv.Coserv, error) {
	var result *coserv.Coserv

	// Broker has keys, meaning the response is signed coserv
	if len(broker.keys) != 0 {
		msg, err := broker.coservclient.RunQueryForSignedResponse(query)
		if err != nil {
			return nil, err
		}

		var coseResp cose.Sign1Message
		if err := coseResp.UnmarshalCBOR(msg); err != nil {
			return nil, err
		}

		alg, err := coseResp.Headers.Protected.Algorithm()
		if err != nil {
			return nil, err
		}

		var verifiers []cose.Verifier
		for _, key := range broker.keys {
			// If key is not following the algorithm, skip it
			if verifier, err := cose.NewVerifier(alg, key); err == nil {
				verifiers = append(verifiers, verifier)
			}
		}

		result, err = coservBuilder.ExtractCoservFromSignedResponse(msg, verifiers)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		result, err = broker.coservclient.RunQueryForUnsignedResponse(query)
		if err != nil {
			return nil, err
		}
	}

	// Checking for CoSERV response expiry
	if result.Results.Expiry.Before(time.Now()) {
		broker.logger.Error("CoSERV response has expired!")
		return nil, errors.New("coserv response has expired")
	}

	// Checking if CoSERV query and query in response are same
	queryCBOR, err := cbor.Marshal(query.Query)
	if err != nil {
		return nil, err
	}

	resultCBOR, err := cbor.Marshal(result.Query)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(resultCBOR, queryCBOR) {
		broker.logger.Error("CoSERV query is not same as query in response!")
		return nil, errors.New("coserv query is not same as query in response")
	}

	return result, nil
}

func (b *Broker) ExecuteCoservQuery(profile string, query string) (*coserv.Coserv, error) {
	if b.coservclient == nil {
		b.logger.Error("CoSERV Client is not Initialized or disabled")
		return nil, ErrFetchFailedStr("coserv client is not initialized, cannot execute coserv query")
	}

	if !slices.Contains(SupportedCoservProfiles, profile) {
		return nil, handler.ErrUnsupported
	}

	var coservQuery coserv.Coserv
	if err := coservQuery.FromBase64Url(query); err != nil {
		return nil, ErrFetchFailed(fmt.Errorf("decoding unsigned CoSERV query: %w", err))
	}

	result, err := fetchCoservResponse(b, &coservQuery)
	if err != nil {
		b.logger.Errorf("Unable to Fetch CoSERV response : %w", err)
		return nil, ErrFetchFailed(fmt.Errorf("unable to Fetch CoSERV response : %w", err))
	}

	if result.Results.AKQ == nil && result.Results.RVQ == nil {
		b.logger.Error("The CoSERV Response is Empty!")
		return nil, handler.ErrNotFound
	}

	b.logger.Debugf("result is: %v", result)
	return result, nil

}

func fetchEndorsements[T any](
	env *comid.Environment,
	buildQuery func(*comid.Environment) (*coserv.Coserv, error),
	extractValues func(*coserv.Coserv) ([]T, error),
	endorsementType string,
	broker *Broker,
) ([]T, error) {

	if broker.coservclient == nil {
		broker.logger.Error("CoSERV Client is not Initialized or disabled")
		return nil, ErrFetchFailedStr(fmt.Sprintf("CoSERV Client is not initialized, cannot fetch %s remotely", endorsementType))
	}

	query, err := buildQuery(env)
	if err != nil {
		broker.logger.Error("Failed to construct CoSERV query")
		return nil, ErrFetchFailed(fmt.Errorf("failed to construct %s query: %w", endorsementType, err))
	}

	result, err := fetchCoservResponse(broker, query)
	if err != nil {
		broker.logger.Errorf("Unable to Fetch CoSERV response : %w", err)
		return nil, ErrFetchFailed(fmt.Errorf("unable to Fetch CoSERV response : %w", err))
	}

	triples, err := extractValues(result)
	if err != nil {
		return nil, err
	}

	broker.logger.Infof("Fetched %s remotely!", endorsementType)
	return triples, nil
}

func (b *Broker) GetKeyTriples(env *comid.Environment, scheme string, exact bool) ([]*comid.KeyTriple, error) {
	inner := func(result *coserv.Coserv) ([]*comid.KeyTriple, error) {
		if result.Results.AKQ == nil {
			b.logger.Error("Trust Anchors not found in cca-coserv-broker store!")
			return nil, handler.ErrNotFound
		}

		keyTriples := make([]*comid.KeyTriple, 0, len(*result.Results.AKQ))
		for _, akq := range *result.Results.AKQ {
			keyTriples = append(keyTriples, akq.AKTriple)
		}

		return keyTriples, nil
	}

	return fetchEndorsements(env, taCoservQuery, inner, "Trust Anchors", b)
}

func (b *Broker) GetValueTriples(env *comid.Environment, scheme string, exact bool) ([]*comid.ValueTriple, error) {
	inner := func(result *coserv.Coserv) ([]*comid.ValueTriple, error) {
		if result.Results.RVQ == nil {
			b.logger.Error("Reference Values not found in cca-coserv-broker store!")
			return nil, handler.ErrNotFound
		}

		valueTriples := make([]*comid.ValueTriple, 0, len(*result.Results.RVQ))
		for _, rvq := range *result.Results.RVQ {
			valueTriples = append(valueTriples, rvq.RVTriple)
		}
		return valueTriples, nil
	}

	return fetchEndorsements(env, rvCoservQuery, inner, "Reference Values", b)
}

func (b *Broker) AddCorimBytes(data []byte, label string, activate bool) error {
	return handler.ErrUnsupported
}

func (b *Broker) SetEndorsementsState(label string, request []byte, setActive bool) error {
	return handler.ErrUnsupported
}

func (b *Broker) Fini() error {
	return nil
}

// Only supporting collected artifacts for now
func taCoservQuery(trustAnchorID *comid.Environment) (*coserv.Coserv, error) {
	envSelector := coserv.NewEnvironmentSelector()
	envSelector.AddInstance(coserv.StatefulInstance{
		Instance:     trustAnchorID.Instance,
		Measurements: nil,
	})
	query, err := coserv.NewEnvironmentQuery(
		coserv.ArtifactTypeTrustAnchors,
		*envSelector,
		coserv.ResultTypeCollectedArtifacts)
	if err != nil {
		return nil, err
	}
	return coserv.NewCoserv(ccaPlatformProfile, *query)
}

func rvCoservQuery(rvID *comid.Environment) (*coserv.Coserv, error) {
	envSelector := coserv.NewEnvironmentSelector()
	envSelector.AddClass(coserv.StatefulClass{
		Class:        rvID.Class,
		Measurements: nil,
	})
	query, err := coserv.NewEnvironmentQuery(
		coserv.ArtifactTypeReferenceValues,
		*envSelector,
		coserv.ResultTypeCollectedArtifacts)
	if err != nil {
		return nil, err
	}
	return coserv.NewCoserv(ccaPlatformProfile, *query)
}
