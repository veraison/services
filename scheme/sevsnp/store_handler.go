// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package sevsnp

import (
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/go-sev-guest/proto/sevsnp"
	"github.com/veraison/corim/comid"
	"github.com/veraison/ratsd/tokens"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/proto"
)

var (
	ErrMissingMeasurement          = errors.New("measurement not found")
	ErrARKDecodeFailure            = errors.New("failed to decode ARK")
	ErrUnsupportedMultipleEvidence = errors.New("unable to process multiple evidence in a single request")
)

// StoreHandler implements the IStoreHandler interface handler for SEVSNP scheme
type StoreHandler struct{}

// GetName returns the name of this StoreHandler instance
func (s StoreHandler) GetName() string {
	return fmt.Sprintf("%s-store-handler", SchemeName)
}

// GetAttestationScheme returns the attestation scheme
func (s StoreHandler) GetAttestationScheme() string {
	return SchemeName
}

// GetSupportedMediaTypes returns the supported media types; no-op for SEVSNP
func (s StoreHandler) GetSupportedMediaTypes() []string {
	return nil
}

// getRefValKey helper to compute RefVal key from CoMID value triple
func getRefValKey(rv comid.ValueTriple, tenantID string) (string, error) {
	m, err := measurementByUintKey(rv, mKeyMeasurement)
	if err != nil {
		return "", err
	}

	if m == nil {
		return "", ErrMissingMeasurement
	}

	d := m.Val.Digests

	u := url.URL{
		Scheme: SchemeName,
		Host:   tenantID,
		Path:   hex.EncodeToString((*d)[0].HashValue),
	}

	return u.String(), nil
}

// SynthKeysFromRefValue constructs SEV-SNP reference value of the form
// "SEVSNP://<tenantID>/<measurement>". The measurement
// is unique to an attester instance and, as such, is
// the best candidate to use as the key.
func (s StoreHandler) SynthKeysFromRefValue(
	tenantID string,
	refValue *handler.Endorsement,
) ([]string, error) {
	var rv comid.ValueTriple

	err := json.Unmarshal(refValue.Attributes, &rv)
	if err != nil {
		return nil, err
	}

	refValKey, err := getRefValKey(rv, tenantID)
	if err != nil {
		return nil, err
	}

	return []string{refValKey}, nil
}

// SynthKeysFromTrustAnchor constructs the SEV-SNP Trust Anchor key. The
// key format is "SEVSNP://<keyname>". For example, "SEV-SNP://ARK-Milan"
//
// AMD's Root Key (ARK) is the only Trust Anchor for SEV-SNP.
//
// The attester supplies all the keys in the certificate chain
// for verification. During verification, the scheme must ensure that
// the ARK in the evidence matches the provisioned Trust Anchor.
func (s StoreHandler) SynthKeysFromTrustAnchor(_ string, ta *handler.Endorsement) ([]string, error) {
	var avk comid.KeyTriple

	err := json.Unmarshal(ta.Attributes, &avk)
	if err != nil {
		return nil, err
	}

	ark := avk.VerifKeys[0]

	keyBlock, _ := pem.Decode([]byte(ark.String()))
	if keyBlock == nil || keyBlock.Type != "CERTIFICATE" {
		return nil, ErrARKDecodeFailure
	}

	cert, err := x509.ParseCertificate(keyBlock.Bytes)
	if err != nil {
		return nil, err
	}

	u := url.URL{
		Scheme: SchemeName,
		Path:   cert.Issuer.CommonName,
	}

	return []string{u.String()}, nil
}

// GetTrustAnchorIDs gets the TA ID from evidence
//
// "auxblob" in the TSM report contains a certificate
// table. Extract ARK from it and construct the TA key.
func (s StoreHandler) GetTrustAnchorIDs(token *proto.AttestationToken) ([]string, error) {
	var (
		tsm       *tokens.TSMReport
		certChain *sevsnp.CertificateChain
		ark       []byte
		cert      *x509.Certificate
		err       error
	)

	if tsm, err = parseAttestationToken(token); err != nil {
		return nil, err
	}

	if certChain, err = parseCertificateChainFromEvidence(tsm); err != nil {
		return nil, err
	}

	if ark, err = readCert(certChain.GetArkCert()); err != nil {
		return nil, fmt.Errorf("can't read ARK to compose TA ID: %w", err)
	}

	if cert, err = x509.ParseCertificate(ark); err != nil {
		return nil, err
	}

	u := url.URL{
		Scheme: SchemeName,
		Path:   cert.Issuer.CommonName,
	}

	return []string{u.String()}, nil
}

// GetRefValueIDs gets the refval key from the claims set. Looks up
// "measurement" using its MKey (641) and construct the refval key.
//
// Reference value key for SEV-SNP is of the form
// "SEVSNP://<tenantID>/<measurement>", as explained
// in SynthKeysFromRefValue.
func (s StoreHandler) GetRefValueIDs(
	tenantID string,
	_ []string,
	claims map[string]interface{},
) ([]string, error) {
	claimsJson, err := json.Marshal(claims)
	if err != nil {
		return nil, err
	}

	extractedComid, err := comidFromJson(claimsJson)
	if err != nil {
		return nil, err
	}

	if len(extractedComid.Triples.ReferenceValues.Values) > 1 {
		return nil, ErrUnsupportedMultipleEvidence
	}

	refValKey, err := getRefValKey(extractedComid.Triples.ReferenceValues.Values[0], tenantID)
	if err != nil {
		return nil, err
	}

	return []string{refValKey}, nil
}

func (s StoreHandler) SynthCoservQueryKeys(tenantID string, query string) ([]string, error) {
	return []string{"TODO"}, nil
}
