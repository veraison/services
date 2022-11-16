// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/x509"
	//"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"github.com/hashicorp/go-plugin"
	nitro_eclave_attestation_document "github.com/veracruz-project/go-nitro-enclave-attestation-document"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme"
	"net/url"
	"time"
)

type Endorsements struct {
}

type Scheme struct{}

func (s Scheme) GetName() string {
	return proto.AttestationFormat_AWS_NITRO.String()
}

func (s Scheme) GetFormat() proto.AttestationFormat {
	return proto.AttestationFormat_AWS_NITRO
}

func (s Scheme) SynthKeysFromSwComponent(tenantID string, swComp *proto.Endorsement) ([]string, error) {

	var return_array []string // intentionally empty, because we have no SW components in our provisioning corim at this time
	return return_array, nil
}

func (s Scheme) SynthKeysFromTrustAnchor(tenantID string, ta *proto.Endorsement) ([]string, error) {
	return []string{nitroTaLookupKey(tenantID)}, nil
}

func (s Scheme) GetSupportedMediaTypes() []string {
	return []string{
		"application/aws-nitro-document",
	}
}

// GetTrustAnchorID returns a string ID used to retrieve a trust anchor
// for this token. The trust anchor may be necessary to validate the
// token and/or extract its claims (if it is encrypted).
func (s Scheme) GetTrustAnchorID(token *proto.AttestationToken) (string, error) {

	return nitroTaLookupKey(token.TenantId), nil
}

// ExtractClaims parses the attestation token and returns claims
// extracted therefrom.
func (s Scheme) ExtractClaims(token *proto.AttestationToken, trustAnchor string) (*scheme.ExtractedClaims, error) {
	return s.extractClaimsImpl(token, trustAnchor, time.Now())
}

/// Same as ExtractClaims, but allows the caller to set an alternate "current time" to allow
/// tests to use saved attestation document data without triggering certificate expiry errors.
/// THIS FUNCTION SHOULD ONLY BE USED IN TESTING
func (s Scheme) ExtractClaimsTest(token *proto.AttestationToken, trustAnchor string, testTime time.Time) (*scheme.ExtractedClaims, error) {
	return s.extractClaimsImpl(token, trustAnchor, testTime)
}

/// Implementation of the functionality for ExtracClaims and ExtracClaimsTest
func (s Scheme) extractClaimsImpl(token *proto.AttestationToken, trustAnchor string, now time.Time) (*scheme.ExtractedClaims, error) {
	ta_unmarshalled := make(map[string]interface{})

	err := json.Unmarshal([]byte(trustAnchor), &ta_unmarshalled)
	if err != nil {
		new_err := fmt.Errorf("ExtractVerifiedClaims call to json.Unmarshall failed:%v", err)
		return nil, new_err
	}
	contents, ok := ta_unmarshalled["attributes"].(map[string]interface{})
	if !ok {
		new_err := fmt.Errorf("scheme-aws-nitro.Scheme.ExtractVerifiedClaims cast of %v to map[string]interface{} failed", ta_unmarshalled["attributes"])
		return nil, new_err
	}

	cert_pem, ok := contents["key"].(string)
	if !ok {
		new_err := fmt.Errorf("scheme-aws-nitro.Scheme.ExtractVerifiedClaims cast of %v to string failed", contents["nitro.iak-pub"])
		return nil, new_err
	}

	// golang standard library pem.Decode function cannot handle PEM data without a header, so I have to add one to make it happy.
	// Yes, this is stupid
	cert_pem = "-----BEGIN CERTIFICATE-----\n" + cert_pem + "\n-----END CERTIFICATE-----\n"
	cert_pem_bytes := []byte(cert_pem)
	cert_block, _ := pem.Decode(cert_pem_bytes)
	if cert_block == nil {
		new_err := fmt.Errorf("scheme-aws-nitro.Scheme.ExtractVerifiedClaims call to pem.Decode failed, but I don't know why")
		return nil, new_err
	}

	cert_der := cert_block.Bytes
	cert, err := x509.ParseCertificate(cert_der)
	if err != nil {
		new_err := fmt.Errorf("scheme-aws-nitro.Scheme.ExtractVerifiedClaims call to x509.ParseCertificate failed:%v", err)
		return nil, new_err
	}

	token_data := token.Data

	var document *nitro_eclave_attestation_document.AttestationDocument
	if flag.Lookup("test.v") == nil {
		document, err = nitro_eclave_attestation_document.AuthenticateDocument(token_data, *cert)
	} else {
		document, err = nitro_eclave_attestation_document.AuthenticateDocumentTest(token_data, *cert, now)
	}
	if err != nil {
		new_err := fmt.Errorf("scheme-aws-nitro.Scheme.ExtractVerifiedClaims call to AuthenticateDocument failed:%v", err)
		return nil, new_err
	}

	var extracted scheme.ExtractedClaims

	claimsSet, err := claimsToMap(document)
	if err != nil {
		new_err := fmt.Errorf("scheme-aws-nitro.Scheme.ExtractVerifiedClaims call to claimsToMap failed:%v", err)
		return nil, new_err
	}
	extracted.ClaimsSet = claimsSet

	return &extracted, nil
}

// AppraiseEvidence evaluates the specified EvidenceContext against
// the specified endorsements, and returns an AttestationResult wrapped
// in an AppraisalContext.
func (s Scheme) AppraiseEvidence(
	ec *proto.EvidenceContext, endorsementsStrings []string,
) (*proto.AppraisalContext, error) {
	appraisalCtx := proto.AppraisalContext{
		Evidence: ec,
		Result:   &proto.AttestationResult{},
	}

	var endorsements []Endorsements

	for i, e := range endorsementsStrings {
		var endorsement Endorsements

		if err := json.Unmarshal([]byte(e), &endorsement); err != nil {
			return nil, fmt.Errorf("could not decode endorsement at index %d: %w", i, err)
		}

		endorsements = append(endorsements, endorsement)
	}

	err := populateAttestationResult(&appraisalCtx, endorsements)

	return &appraisalCtx, err
}

// ValidateEvidenceIntegrity verifies the structural integrity and validity of the
// token. The exact checks performed are scheme-specific, but they
// would typically involve, at the least, verifying the token's
// signature using the provided trust anchor. If the validation fails,
// an error detailing what went wrong is returned.
// TODO(setrofim): no distinction is currently made between validation
// failing due to an internal error, and it failing due to bad input
// (i.e. signature not matching).
func (s Scheme) ValidateEvidenceIntegrity(
	token *proto.AttestationToken,
	trustAnchor string,
	endorsementsStrings []string,
) error {
	return s.validateEvidenceIntegrityImpl(token, trustAnchor, endorsementsStrings, time.Now())
}

/// Same as ValidateEvidenceIntegrity, but allows the caller to set an alternate "current time" to allow
/// tests to use saved attestation document data without triggering certificate expiry errors.
/// THIS FUNCTION SHOULD ONLY BE USED IN TESTING
func (s Scheme) ValidateEvidenceIntegrityTest(
	token *proto.AttestationToken,
	trustAnchor string,
	endorsementsStrings []string,
	testTime time.Time,
) error {
	return s.validateEvidenceIntegrityImpl(token, trustAnchor, endorsementsStrings, testTime)
}

func (s Scheme) validateEvidenceIntegrityImpl(token *proto.AttestationToken,
	trustAnchor string,
	endorsementsStrings []string,
	now time.Time,
) error {

	ta_unmarshalled := make(map[string]interface{})

	err := json.Unmarshal([]byte(trustAnchor), &ta_unmarshalled)
	if err != nil {
		new_err := fmt.Errorf("ValidateEvidenceIntegrityImpl call to json.Unmarshall failed:%v", err)
		return new_err
	}
	contents, ok := ta_unmarshalled["attributes"].(map[string]interface{})
	if !ok {
		new_err := fmt.Errorf("scheme-aws-nitro.Scheme.ValidateEvidenceIntegrityImpl cast of %v to map[string]interface{} failed", ta_unmarshalled["attributes"])
		return new_err
	}

	cert_pem, ok := contents["key"].(string)
	if !ok {
		new_err := fmt.Errorf("scheme-aws-nitro.Scheme.ValidateEvidenceIntegrityImpl cast of %v to string failed", contents["nitro.iak-pub"])
		return new_err
	}

	// golang standard library pem.Decode function cannot handle PEM data without a header, so I have to add one to make it happy.
	// Yes, this is stupid
	cert_pem = "-----BEGIN CERTIFICATE-----\n" + cert_pem + "\n-----END CERTIFICATE-----\n"
	cert_pem_bytes := []byte(cert_pem)
	cert_block, _ := pem.Decode(cert_pem_bytes)
	if cert_block == nil {
		new_err := fmt.Errorf("scheme-aws-nitro.Scheme.ValidateEvidenceIntegrityImpl call to pem.Decode failed, but I don't know why")
		return new_err
	}

	cert_der := cert_block.Bytes
	cert, err := x509.ParseCertificate(cert_der)
	if err != nil {
		new_err := fmt.Errorf("scheme-aws-nitro.Scheme.ValidateEvidenceIntegrityImpl call to x509.ParseCertificate failed:%v", err)
		return new_err
	}

	// token_data, err := base64.StdEncoding.DecodeString(string(token.Data))
	// if err != nil {
	// 	new_err := fmt.Errorf("scheme-aws-nitro.Scheme.ValidateEvidenceIntegrityImpl call to base64.StdEncoding.DecodeString failed:%v", err)
	// 	return nil, new_err
	// }
	token_data := token.Data

	if flag.Lookup("test.v") == nil {
		_, err = nitro_eclave_attestation_document.AuthenticateDocument(token_data, *cert)
	} else {
		_, err = nitro_eclave_attestation_document.AuthenticateDocumentTest(token_data, *cert, now)
	}
	if err != nil {
		new_err := fmt.Errorf("scheme-aws-nitro.Scheme.ValidateEvidenceIntegrityImpl call to AuthenticateDocument failed:%v", err)
		return new_err
	}
	return nil
}

func claimsToMap(doc *nitro_eclave_attestation_document.AttestationDocument) (out map[string]interface{}, err error) {
	out = make(map[string]interface{})
	for index, this_pcr := range doc.PCRs {
		var key = fmt.Sprintf("PCR%v", index)
		out[key] = this_pcr
	}
	out["user_data"] = doc.User_Data
	out["nonce"] = doc.Nonce

	return out, nil
}

func populateAttestationResult(appraisalCtx *proto.AppraisalContext, endorsements []Endorsements) error {
	tv := proto.TrustVector{
		InstanceIdentity: int32(proto.ARStatus_NO_CLAIM),
		Configuration:    int32(proto.ARStatus_NO_CLAIM),
		Executables:      int32(proto.ARStatus_NO_CLAIM),
		FileSystem:       int32(proto.ARStatus_NO_CLAIM),
		Hardware:         int32(proto.ARStatus_NO_CLAIM),
		RuntimeOpaque:    int32(proto.ARStatus_NO_CLAIM),
		StorageOpaque:    int32(proto.ARStatus_NO_CLAIM),
		SourcedData:      int32(proto.ARStatus_NO_CLAIM),
	}

	// once the signature on the token is verified, we can claim the HW is
	// authentic
	tv.Hardware = int32(proto.ARStatus_HW_AFFIRMING)

	appraisalCtx.Result.TrustVector = &tv

	appraisalCtx.Result.Status = proto.TrustTier_AFFIRMING

	appraisalCtx.Result.ProcessedEvidence = appraisalCtx.Evidence.Evidence

	return nil
}

func nitroTaLookupKey(tenantID string) string {

	u := url.URL{
		Scheme: proto.AttestationFormat_AWS_NITRO.String(),
		Host:   tenantID,
		Path:   "/",
	}

	return u.String()
}

func main() {
	var handshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "VERAISON_PLUGIN",
		MagicCookieValue: "VERAISON",
	}

	var pluginMap = map[string]plugin.Plugin{
		"scheme": &scheme.Plugin{
			Impl: &Scheme{},
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
