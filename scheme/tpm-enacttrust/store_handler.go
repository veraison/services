// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package tpm_enacttrust

import (
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

func (s StoreHandler) SynthKeysFromRefValue(
	tenantID string,
	swComp *handler.Endorsement,
) ([]string, error) {
	return synthKeysFromAttrs("software component", tenantID, swComp.Attributes)
}

func (s StoreHandler) SynthKeysFromTrustAnchor(tenantID string, ta *handler.Endorsement) ([]string, error) {
	return synthKeysFromAttrs("trust anchor", tenantID, ta.Attributes)
}
