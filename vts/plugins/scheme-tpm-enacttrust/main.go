// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	tpm2 "github.com/google/go-tpm/tpm2"
	uuid "github.com/google/uuid"
	plugin "github.com/hashicorp/go-plugin"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme"
	"github.com/veraison/services/vts/plugins/common"
	structpb "google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Scheme struct{}

func (s Scheme) GetName() string {
	return proto.AttestationFormat_TPM_ENACTTRUST.String()
}

func (s Scheme) GetFormat() proto.AttestationFormat {
	return proto.AttestationFormat_TPM_ENACTTRUST
}

func (s Scheme) GetSupportedMediaTypes() []string {
	return []string{
		"application/vnd.enacttrust.tpm-evidence",
	}
}

func (s Scheme) SynthKeysFromSwComponent(tenantID string, swComp *proto.Endorsement) ([]string, error) {
	return synthKeysFromParts("software component", tenantID, swComp.GetAttributes())
}

func (s Scheme) SynthKeysFromTrustAnchor(tenantID string, ta *proto.Endorsement) ([]string, error) {
	return synthKeysFromParts("trust anchor", tenantID, ta.GetAttributes())
}

func (s Scheme) GetTrustAnchorID(token *proto.AttestationToken) (string, error) {
	if token.Format != proto.AttestationFormat_TPM_ENACTTRUST {
		return "", fmt.Errorf("wrong format: expect %q, but found %q",
			proto.AttestationFormat_TPM_ENACTTRUST.String(),
			token.Format.String(),
		)
	}

	var decoded Token

	if err := decoded.Decode(token.Data); err != nil {
		return "", err
	}

	nodeID, err := uuid.FromBytes(decoded.AttestationData.ExtraData)
	if err != nil {
		return "", fmt.Errorf("could not decode node-id: %v", err)
	}

	return tpmEnactTrustLookupKey(token.TenantId, nodeID.String()), nil
}

func (s Scheme) ExtractVerifiedClaims(token *proto.AttestationToken, trustAnchor string) (*scheme.ExtractedClaims, error) {
	if token.Format != proto.AttestationFormat_TPM_ENACTTRUST {
		return nil, fmt.Errorf("wrong format: expect %q, but found %q",
			proto.AttestationFormat_TPM_ENACTTRUST.String(),
			token.Format.String(),
		)
	}

	var decoded Token

	if err := decoded.Decode(token.Data); err != nil {
		return nil, fmt.Errorf("could not decode token: %w", err)
	}

	pubKey, err := parseKey(trustAnchor)
	if err != nil {
		return nil, fmt.Errorf("could not parse trust anchor: %w", err)
	}

	if err = decoded.VerifySignature(pubKey); err != nil {
		return nil, fmt.Errorf("could not verify token signature: %w", err)
	}

	if decoded.AttestationData.Type != tpm2.TagAttestQuote {
		return nil, fmt.Errorf("wrong TPMS_ATTEST type: want %d, got %d",
			tpm2.TagAttestQuote, decoded.AttestationData.Type)
	}

	var pcrs []int64
	for _, pcr := range decoded.AttestationData.AttestedQuoteInfo.PCRSelection.PCRs {
		pcrs = append(pcrs, int64(pcr))
	}

	evidence := scheme.NewExtractedClaims()
	evidence.ClaimsSet["pcr-selection"] = pcrs
	evidence.ClaimsSet["hash-algorithm"] = int64(decoded.AttestationData.AttestedQuoteInfo.PCRSelection.Hash)
	evidence.ClaimsSet["pcr-digest"] = []byte(decoded.AttestationData.AttestedQuoteInfo.PCRDigest)

	nodeID, err := uuid.FromBytes(decoded.AttestationData.ExtraData)
	if err != nil {
		return nil, fmt.Errorf("could not decode node-id: %w", err)
	}
	evidence.SoftwareID = tpmEnactTrustLookupKey(token.TenantId, nodeID.String())

	return evidence, nil
}

func initAppraisalContext(ec *proto.EvidenceContext) *proto.AppraisalContext {
	return &proto.AppraisalContext{
		Evidence: ec,
		Result: &proto.AttestationResult{
			Status: proto.AR_Status_FAILURE,
			TrustVector: &proto.TrustVector{
				SoftwareIntegrity:    proto.AR_Status_FAILURE,
				HardwareAuthenticity: proto.AR_Status_UNKNOWN,
				SoftwareUpToDateness: proto.AR_Status_UNKNOWN,
				ConfigIntegrity:      proto.AR_Status_UNKNOWN,
				RuntimeIntegrity:     proto.AR_Status_UNKNOWN,
				CertificationStatus:  proto.AR_Status_UNKNOWN,
			},
			Timestamp:         timestamppb.Now(),
			ProcessedEvidence: ec.Evidence,
		},
	}
}

func (s Scheme) AppraiseEvidence(ec *proto.EvidenceContext, endorsementStrings []string) (*proto.AppraisalContext, error) {
	appraisalCtx := initAppraisalContext(ec)

	digestValue, ok := ec.Evidence.AsMap()["pcr-digest"]
	if !ok {
		return appraisalCtx, fmt.Errorf("evidence does not contain %q entry", "pcr-digest")
	}

	evidenceDigest, ok := digestValue.(string)
	if !ok {
		return appraisalCtx, fmt.Errorf("wrong type value %q entry; expected string but found %T", "pcr-digest", digestValue)
	}

	var endorsements Endorsements
	if err := endorsements.Populate(endorsementStrings); err != nil {
		return appraisalCtx, err
	}

	if endorsements.Digest == evidenceDigest {
		appraisalCtx.Result.TrustVector.SoftwareIntegrity = proto.AR_Status_SUCCESS
		appraisalCtx.Result.Status = proto.AR_Status_SUCCESS
	}

	return appraisalCtx, nil
}

func synthKeysFromParts(scope, tenantID string, parts *structpb.Struct) ([]string, error) {
	var (
		nodeID string
		fields map[string]*structpb.Value
		err    error
	)

	fields, err = common.GetFieldsFromParts(parts)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize %s abs-path: %w", scope, err)
	}

	nodeID, err = common.GetMandatoryPathSegment("enacttrust-tpm.node-id", fields)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize %s abs-path: %w", scope, err)
	}

	return []string{tpmEnactTrustLookupKey(tenantID, nodeID)}, nil
}

func parseKey(keyString string) (*ecdsa.PublicKey, error) {
	buf, err := base64.StdEncoding.DecodeString(keyString)
	if err != nil {
		return nil, err
	}

	key, err := x509.ParsePKIXPublicKey(buf)
	if err != nil {
		return nil, fmt.Errorf("could not parse public key: %v", err)
	}

	ret, ok := key.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("could not extract EC public key; got [%T]: %v", key, err)
	}

	return ret, nil
}

func tpmEnactTrustLookupKey(tenantID, nodeID string) string {
	absPath := []string{nodeID}

	u := url.URL{
		Scheme: proto.AttestationFormat_TPM_ENACTTRUST.String(),
		Host:   tenantID,
		Path:   strings.Join(absPath, "/"),
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
