// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package tpm_enacttrust

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/veraison/services/handler"
	"github.com/veraison/services/proto"
)

type StoreHandler struct {
}

func (s StoreHandler) GetName() string {
	return "tpm-enacttrust-store-handler"
}

func (s StoreHandler) GetAttestationScheme() string {
	return SchemeName
}

func (s StoreHandler) GetSupportedMediaTypes() []string {
	return nil
}

func (s StoreHandler) GetTrustAnchorIDs(token *proto.AttestationToken) ([]string, error) {
	supported := false
	for _, mt := range EvidenceMediaTypes {
		if token.MediaType == mt {
			supported = true
			break
		}
	}

	if !supported {
		err := handler.BadEvidence(
			"wrong media type: expect %q, but found %q",
			strings.Join(EvidenceMediaTypes, ", "),
			token.MediaType,
		)
		return []string{""}, err
	}

	var decoded Token

	if err := decoded.Decode(token.Data); err != nil {
		return nil, handler.BadEvidence(err)
	}

	return []string{tpmEnactTrustLookupKey(token.TenantId, decoded.NodeId.String())}, nil
}

func (s StoreHandler) GetRefValueIDs(
	tenantID string,
	trustAnchors []string,
	claims map[string]interface{},
) ([]string, error) {
	nodeID, ok := claims["node-id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid node-id value: %v", claims["node-id"])
	}

	return []string{tpmEnactTrustLookupKey(tenantID, nodeID)}, nil
}

func (s StoreHandler) SynthKeysFromRefValue(
	tenantID string,
	swComp *handler.Endorsement,
) ([]string, error) {
	return synthKeysFromAttrs("software component", tenantID, swComp.Attributes)
}

func (s StoreHandler) SynthKeysFromTrustAnchor(tenantID string, ta *handler.Endorsement) ([]string, error) {
	return synthKeysFromAttrs("trust anchor", tenantID, ta.Attributes)
}

func synthKeysFromAttrs(scope string, tenantID string, attr json.RawMessage) ([]string, error) {
	var (
		nodeID string
		err    error
	)

	switch scope {
	case "software component":
		var att RefValAttr
		if err = json.Unmarshal(attr, &att); err != nil {
			return nil, fmt.Errorf("unable to extract sw component: %w", err)
		}
		nodeID = att.NodeID
	case "trust anchor":
		var att TaAttr
		if err = json.Unmarshal(attr, &att); err != nil {
			return nil, fmt.Errorf("unable to extract trust anchor: %w", err)
		}
		nodeID = att.NodeID
	default:
		return nil, fmt.Errorf("invalid scope: %s", scope)
	}

	return []string{tpmEnactTrustLookupKey(tenantID, nodeID)}, nil
}

func tpmEnactTrustLookupKey(tenantID, nodeID string) string {
	absPath := []string{nodeID}

	u := url.URL{
		Scheme: SchemeName,
		Host:   tenantID,
		Path:   strings.Join(absPath, "/"),
	}

	return u.String()
}

func (s StoreHandler) SynthCoservQueryKeys(tenantID string, query string) ([]string, error) {
	return []string{"TODO"}, nil
}
