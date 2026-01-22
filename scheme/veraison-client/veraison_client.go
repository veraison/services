// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package veraisonclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/veraison/apiclient/verification"
)

type ComponentVerifierClientHandler struct{}

func (ComponentVerifierClientHandler) GetName() string {
	return "veraison-client-handler"
}

func (ComponentVerifierClientHandler) GetAttestationScheme() string {
	return SchemeName
}

func (ComponentVerifierClientHandler) GetSupportedMediaTypes() []string {
	return VeraisonClientMediaTypes
}

type ClientConfig struct {
	DiscoveryURL string   `json:"url"`
	CACerts      []string `json:"ca_certs,omitempty"`
	Insecure     bool     `json:"insecure,omitempty"`
	crURL        string   // the challenge-response URL is discovered dynamically
}

func unpackConfig(clientCfg []byte) (*ClientConfig, error) {
	var cfg ClientConfig
	if err := json.Unmarshal(clientCfg, &cfg); err != nil {
		return nil, fmt.Errorf("decoding JSON-encoded clientCfg: %w", err)
	}

	if cfg.DiscoveryURL == "" {
		return nil, errors.New("missing mandatory URL")
	}

	return &cfg, nil
}

func discover(cfg *ClientConfig, vc IVeraisonDiscoveryClient) (*jwk.Key, string, error) {
	if err := vc.SetDiscoveryURI(cfg.DiscoveryURL); err != nil {
		return nil, "", fmt.Errorf("failed to set discovery URI: %w", err)
	}

	if len(cfg.CACerts) > 0 {
		if err := vc.SetCerts(cfg.CACerts); err != nil {
			return nil, "", fmt.Errorf("failed to set CA certs: %w", err)
		}
	}

	if cfg.Insecure {
		vc.SetIsInsecure()
	}

	discoveryObject, err := vc.Run()
	if err != nil {
		return nil, "", fmt.Errorf("failed to run discovery: %w", err)
	}

	verificationKey, err := jwk.ParseKey(discoveryObject.PublicKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse verification key: %w", err)
	}

	// construct the c-r URL from base
	crURL, err := constructChallengeResponseURL(
		cfg.DiscoveryURL,
		discoveryObject.ApiEndpoints["newChallengeResponseSession"],
	)

	if err != nil {
		return nil, "", fmt.Errorf("failed to construct challenge-response URL: %w", err)
	}

	return &verificationKey, crURL, nil
}

func constructChallengeResponseURL(baseURL, crEndpoint string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parsing base URL: %w", err)
	}

	if crEndpoint == "" {
		return "", errors.New("challenge-response endpoint is empty")
	}

	ref, err := url.Parse(crEndpoint)
	if err != nil {
		return "", fmt.Errorf("parsing challenge-response endpoint: %w", err)
	}

	crURL := base.ResolveReference(ref)

	return crURL.String(), nil
}

// remoteAppraisal performs the challenge-response protocol in RP mode.
func remoteAppraisal(
	evidence []byte,
	mediaType string,
	nonce []byte,
	cfg *ClientConfig,
	vc IVeraisonChallengeResponseClient,
) ([]byte, error) {
	if err := vc.SetSessionURI(cfg.crURL); err != nil {
		return nil, fmt.Errorf("failed to set session URI: %w", err)
	}

	vc.SetDeleteSession(true)
	if len(cfg.CACerts) > 0 {
		vc.SetCerts(cfg.CACerts)
	}

	vc.SetIsInsecure(cfg.Insecure)

	if err := vc.SetNonce(nonce); err != nil {
		return nil, fmt.Errorf("failed to set nonce: %w", err)
	}

	if err := vc.SetEvidenceBuilder(
		verification.NewStaticEvidenceBuilder(evidence, mediaType),
	); err != nil {
		return nil, fmt.Errorf("failed to set evidence builder: %w", err)
	}

	arBytes, err := vc.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run challenge-response client: %w", err)
	}

	return arBytes, nil
}

func verifyAttestationResult(arBytes []byte, verificationKey *jwk.Key) ([]byte, error) {
	// Verify the signature of the attestation result.
	//
	// TODO(tho): remove hard-coded alg.  Instead, access the signature
	//            algorithm from the JWT header.
	t, err := jwt.Parse(arBytes, jwt.WithKey(jwa.ES256, *verificationKey), jwt.WithValidate(true))
	if err != nil {
		return nil, fmt.Errorf("verifying attestation result signature: %w", err)
	}

	// Extract the EAR appraisal(s) from the attestation result.
	submods, ok := t.Get("submods")
	if !ok {
		return nil, errors.New("no appraisals in EAR")
	}

	appraisals, err := json.Marshal(submods)
	if err != nil {
		return nil, fmt.Errorf("marshaling appraisals to JSON: %w", err)
	}

	return appraisals, nil
}

func appraiseComponentEvidence(
	evidence []byte,
	mediaType string,
	nonce []byte,
	clientCfg []byte,
	dcc IVeraisonDiscoveryClient,
	crc IVeraisonChallengeResponseClient,
) ([]byte, error) {
	// 1. Unpack the clientCfg to obtain the verifier transport and trust
	//    settings.
	cfg, err := unpackConfig(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("invalid client configuration: %w", err)
	}

	// 2. Get the verifier's public key and the C-R endpoint by querying the
	// well-known interface.
	verificationKey, crURL, err := discover(cfg, dcc)
	if err != nil {
		return nil, fmt.Errorf("failed to get verifier's public key: %w", err)
	}

	// store the discovered C-R session URL in cfg for use in remoteAppraisal
	cfg.crURL = crURL

	// 3. Initiate a challenge-response session in RP mode with the configured
	//    verifier, supplying the component evidence and nonce.
	// 4. Obtain an EAR from the verifier.
	arBytes, err := remoteAppraisal(evidence, mediaType, nonce, cfg, crc)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain EAR from verifier: %w", err)
	}

	// 5. Verify the signature of the EAR.
	appraisal, err := verifyAttestationResult(arBytes, verificationKey)
	if err != nil {
		return nil, fmt.Errorf("failed to verify attestation result: %w", err)
	}

	// 6. Return the EAR appraisal to the CE Handler.
	return appraisal, nil
}

func (ComponentVerifierClientHandler) AppraiseComponentEvidence(
	evidence []byte,
	mediaType string,
	nonce []byte,
	clientCfg []byte,
) ([]byte, error) {
	return appraiseComponentEvidence(
		evidence,
		mediaType,
		nonce,
		clientCfg,
		&verification.DiscoveryConfig{},
		&verification.ChallengeResponseConfig{},
	)
}
