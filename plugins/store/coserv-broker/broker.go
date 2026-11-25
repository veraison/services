// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package coservbroker

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
	coservapi "github.com/veraison/apiclient/coserv"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/coserv"
	cose "github.com/veraison/go-cose"
	"github.com/veraison/services/pkg/endorsement"
	"github.com/veraison/services/pkg/fault"
	"github.com/veraison/services/pkg/log"
	"github.com/veraison/services/pkg/plugin"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

const PluginName = "coserv-broker"

var Capabilities = []endorsement.Capability{
	endorsement.ExecuteCoservQuery,
}

var (
	CoservMediaTypes        []string
	SupportedCoservProfiles []string
)

var ErrFetchFailed = errors.New("coserv-broker fetch failed")

func FetchFailed(reason error) error {
	return fmt.Errorf("%w: %w", ErrFetchFailed, reason)
}

type SchemeConfig struct {
	CoservURL string   `yaml:"coserv-url"`
	Insecure  bool     `yaml:"client-tls-insecure"`
	Certs     []string `yaml:"client-certs"`
	Unsigned  bool     `yaml:"unsigned-coserv"`
	Profile   string   `yaml:"profile"`
}

type ClientConfig struct {
	coservclient *coservapi.QueryConfig
	keys         []crypto.PublicKey
	profile      string
}

type Broker struct {
	logger    *zap.SugaredLogger
	clientcfg map[string]ClientConfig
}

func NewBroker() *Broker {
	return &Broker{
		logger:    log.Named(PluginName),
		clientcfg: make(map[string]ClientConfig),
	}
}

func (b *Broker) Init(params *plugin.Parameters) error {
	if params == nil {
		return errors.New("parameters are required for coserv-broker")
	}

	enablecache, err := params.DefaultGetBool("client-caching", true)
	if err != nil {
		return badConfig("client-caching config is not a bool!")
	}

	schemes, err := params.DefaultGetParameters("schemes", nil)
	if err != nil {
		return badConfig("schemes config is incorrectly set!")
	}
	if schemes == nil {
		return badConfig("no schemes are added to coserv broker config!")
	}

	for scheme, cfg := range schemes.Map() {
		coservDiscovery := &coservapi.DiscoveryConfig{}
		coservclient := &coservapi.QueryConfig{}
		clientcfg := ClientConfig{}

		raw, err := yaml.Marshal(cfg)
		if err != nil {
			return badConfig("unable to marshal scheme config: %w", err)
		}

		var schemecfg SchemeConfig
		if err := yaml.Unmarshal(raw, &schemecfg); err != nil {
			return badConfig("unable to parse scheme config: %w", err)
		}

		if err := coservDiscovery.SetDiscoveryURI(schemecfg.CoservURL); err != nil {
			return badConfig("failed to set coserv-url in discovery client! : %w", err)
		}

		coservDiscovery.SetIsInsecure(schemecfg.Insecure)
		coservclient.SetIsInsecure(schemecfg.Insecure)

		if len(schemecfg.Certs) > 0 {
			if err := coservDiscovery.SetCerts(schemecfg.Certs); err != nil {
				return badConfig("failed to add certs in discovery client! : %w", err)
			}
			if err := coservclient.SetCerts(schemecfg.Certs); err != nil {
				return badConfig("failed to add certs in coserv client! : %w", err)
			}
		}

		coservDiscovery.EnableCache(enablecache)
		coservclient.EnableCache(enablecache)

		discoveryResult, err := coservDiscovery.Run()
		if err != nil {
			return badConfig("failed to fetch discovery document! : %w", err)
		}

		b.logger.Info("Successfully fetched Discovery Document")

		if err := coservclient.SetRequestResponseURI(discoveryResult.QueryEndpointURL); err != nil {
			return badConfig("failed to set coserv query endpoint url in coserv client! : %w", err)
		}

		if schemecfg.Profile == "" {
			return badConfig("profile config is empty!")
		}

		var mediatype string
		if schemecfg.Unsigned {
			mediatype = fmt.Sprintf(
				`application/coserv+cbor; profile=%q`,
				schemecfg.Profile,
			)
		} else {
			mediatype = fmt.Sprintf(
				`application/coserv+cose; profile=%q`,
				schemecfg.Profile,
			)
		}

		hasCapability := false
		for _, capability := range discoveryResult.CapabilitiesList {
			if capability.MediaType == mediatype {
				hasCapability = true
				break
			}
		}
		if !hasCapability {
			return badConfig(
				"remote coserv service is unable to serve %s mediatype response",
				mediatype,
			)
		}

		if !schemecfg.Unsigned {
			switch {
			case len(discoveryResult.VerificationKeyJwk) > 0:
				for _, key := range discoveryResult.VerificationKeyJwk {
					var jwk jose.JSONWebKey
					if err := jwk.UnmarshalJSON(key); err != nil {
						return badConfig(
							"failed to parse JWK key in discovery document! : %w",
							err,
						)
					}
					if pb, ok := jwk.Key.(crypto.PublicKey); ok {
						clientcfg.keys = append(clientcfg.keys, pb)
					} else {
						return badConfig(
							"unable to convert JWK key to crypto PublicKey format!",
						)
					}
				}
			case len(discoveryResult.VerificationKeyCose) > 0:
				for _, key := range discoveryResult.VerificationKeyCose {
					var coseKey cose.Key
					if err := coseKey.UnmarshalCBOR(key); err != nil {
						return badConfig(
							"failed to parse COSE key in discovery document! : %w",
							err,
						)
					}
					if pb, err := coseKey.PublicKey(); err == nil {
						clientcfg.keys = append(clientcfg.keys, pb)
					} else {
						return badConfig(
							"unable to convert COSE key to crypto PublicKey format! : %w",
							err,
						)
					}
				}
			default:
				return badConfig(
					"signed coserv is configured for result but no keys found in discovery document!",
				)
			}
		}
		clientcfg.coservclient = coservclient
		clientcfg.profile = schemecfg.Profile
		SupportedCoservProfiles = append(SupportedCoservProfiles, schemecfg.Profile)
		CoservMediaTypes = append(CoservMediaTypes, mediatype)
		b.clientcfg[scheme] = clientcfg
	}

	return nil
}

func (b *Broker) GetName() string {
	return PluginName
}

// Fetches CoSERV response from the remote service based on CoSERV profile and base64url query
func (b *Broker) ExecuteCoservQuery(profile string, query string) (*coserv.Coserv, error) {
	if !slices.Contains(SupportedCoservProfiles, profile) {
		return nil, fault.ErrNotSupported
	}

	for _, clientcfg := range b.clientcfg {
		// If profile doesn't match then skip
		if clientcfg.profile != profile {
			continue
		}

		var coservQuery coserv.Coserv
		if err := coservQuery.FromBase64Url(query); err != nil {
			return nil, FetchFailed(fmt.Errorf("decoding unsigned CoSERV query: %w", err))
		}

		result, err := fetchCoservResponse(&clientcfg, &coservQuery, b.logger)
		if err != nil {
			b.logger.Errorf("Unable to Fetch CoSERV response : %w", err)
			return nil, FetchFailed(fmt.Errorf("unable to Fetch CoSERV response : %w", err))
		}

		if result.Results.AKQ == nil && result.Results.RVQ == nil {
			b.logger.Error("The CoSERV Response is Empty!")
			return nil, fault.ErrNotFound
		}

		b.logger.Debugf("result is: %v", result)
		return result, nil
	}

	return nil, fault.ErrNotSupported
}

func fetchEndorsements[T any](
	env *comid.Environment,
	label string,
	extractValues func(*coserv.Coserv) ([]T, error),
	endorsementType coserv.ArtifactType,
	broker *Broker,
) ([]T, error) {

	_, scheme, found := strings.Cut(label, "/")
	if !found {
		broker.logger.Error("Scheme not found in Label!")
		return nil, FetchFailed(errors.New("scheme not found in label"))
	}
	scheme = strings.ToLower(scheme)

	if clientcfg, ok := broker.clientcfg[scheme]; !ok {
		broker.logger.Errorf("CoSERV Client is not Initialized or disabled for scheme: %s", scheme)
		return nil, FetchFailed(fmt.Errorf("CoSERV Client is not initialized for %s scheme", scheme))
	} else {
		query, err := coservQuery(env, clientcfg.profile, endorsementType)
		if err != nil {
			broker.logger.Error("Failed to construct CoSERV query")
			return nil, FetchFailed(fmt.Errorf("failed to construct %s query: %w", endorsementType, err))
		}

		result, err := fetchCoservResponse(&clientcfg, query, broker.logger)
		if err != nil {
			broker.logger.Errorf("Unable to Fetch CoSERV response : %w", err)
			return nil, FetchFailed(fmt.Errorf("unable to Fetch CoSERV response : %w", err))
		}

		triples, err := extractValues(result)
		if err != nil {
			return nil, err
		}

		broker.logger.Infof("Fetched %s remotely!", endorsementType)
		return triples, nil
	}
}

//nolint:dupl
func (b *Broker) GetKeyTriples(env *comid.Environment, label string, exact bool) ([]*comid.KeyTriple, error) {
	inner := func(result *coserv.Coserv) ([]*comid.KeyTriple, error) {
		if result.Results.AKQ == nil {
			b.logger.Error("Trust Anchors not found in cca-coserv-broker store!")
			return nil, fault.ErrNotFound
		}

		keyTriples := make([]*comid.KeyTriple, 0, len(*result.Results.AKQ))
		for _, akq := range *result.Results.AKQ {
			keyTriples = append(keyTriples, akq.AKTriple)
		}

		return keyTriples, nil
	}

	return fetchEndorsements(env, label, inner, coserv.ArtifactTypeTrustAnchors, b)
}

//nolint:dupl
func (b *Broker) GetValueTriples(env *comid.Environment, label string, exact bool) ([]*comid.ValueTriple, error) {
	inner := func(result *coserv.Coserv) ([]*comid.ValueTriple, error) {
		if result.Results.RVQ == nil {
			b.logger.Error("Reference Values not found in cca-coserv-broker store!")
			return nil, fault.ErrNotFound
		}

		valueTriples := make([]*comid.ValueTriple, 0, len(*result.Results.RVQ))
		for _, rvq := range *result.Results.RVQ {
			valueTriples = append(valueTriples, rvq.RVTriple)
		}
		return valueTriples, nil
	}

	return fetchEndorsements(env, label, inner, coserv.ArtifactTypeReferenceValues, b)
}

func badConfig(args ...any) error {
	return fault.BadConfig("coserv-broker", args...)
}

// Fetches and validates CoSERV response
func fetchCoservResponse(
	clientcfg *ClientConfig,
	query *coserv.Coserv,
	logger *zap.SugaredLogger,
) (*coserv.Coserv, error) {
	var result *coserv.Coserv

	if clientcfg.coservclient == nil {
		logger.Error("CoSERV Client is nil!")
		return nil, errors.New("coserv client is nil")
	}

	// Broker has keys, meaning the response is signed coserv
	if len(clientcfg.keys) != 0 {
		msg, err := clientcfg.coservclient.RunQueryForSignedResponse(query)
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
		for _, key := range clientcfg.keys {
			// If key is not following the algorithm, skip it
			if verifier, err := cose.NewVerifier(alg, key); err == nil {
				verifiers = append(verifiers, verifier)
			}
		}

		result, err = coservapi.ExtractCoservFromSignedResponse(msg, verifiers)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		result, err = clientcfg.coservclient.RunQueryForUnsignedResponse(query)
		if err != nil {
			return nil, err
		}
	}

	// Checking for CoSERV response expiry
	if result.Results.Expiry.Before(time.Now()) {
		logger.Error("CoSERV response has expired!")
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
		logger.Error("CoSERV query is not same as query in response!")
		return nil, errors.New("coserv query is not same as query in response")
	}

	return result, nil
}

// Only supporting collected artifacts for now
func coservQuery(env *comid.Environment, profile string, artifactType coserv.ArtifactType) (*coserv.Coserv, error) {
	envSelector := coserv.NewEnvironmentSelector()
	switch {
	case env.Instance != nil:
		envSelector.AddInstance(coserv.StatefulInstance{
			Instance:     env.Instance,
			Measurements: nil,
		})
	case env.Group != nil:
		envSelector.AddGroup(coserv.StatefulGroup{
			Group:        env.Group,
			Measurements: nil,
		})
	case env.Class != nil:
		envSelector.AddClass(coserv.StatefulClass{
			Class:        env.Class,
			Measurements: nil,
		})
	default:
		return nil, fmt.Errorf("no environment selector for {%s} found", artifactType)
	}
	query, err := coserv.NewEnvironmentQuery(
		artifactType,
		*envSelector,
		coserv.ResultTypeCollectedArtifacts)
	if err != nil {
		return nil, err
	}
	return coserv.NewCoserv(profile, *query)
}
