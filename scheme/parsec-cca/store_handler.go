// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_cca

import (
	"fmt"

	parsec_cca "github.com/veraison/parsec/cca"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme/common/arm"
)

type StoreHandler struct{}

func (s StoreHandler) GetName() string {
	return "parsec-cca-store-handler"
}

func (s StoreHandler) GetAttestationScheme() string {
	return SchemeName
}

func (s StoreHandler) GetSupportedMediaTypes() []string {
	return nil
}

func (s StoreHandler) SynthKeysFromRefValue(
	tenantID string,
	refVal *handler.Endorsement,
) ([]string, error) {

	return arm.SynthKeysForPlatform(SchemeName, tenantID, refVal)
}

func (s StoreHandler) SynthKeysFromTrustAnchor(tenantID string, ta *handler.Endorsement) ([]string, error) {

	return arm.SynthKeysFromTrustAnchors(SchemeName, tenantID, ta)
}

func (s StoreHandler) GetTrustAnchorIDs(token *proto.AttestationToken) ([]string, error) {
	var evidence parsec_cca.Evidence

	err := evidence.FromCBOR(token.Data)
	if err != nil {
		return []string{""}, handler.BadEvidence(err)
	}
	claims := evidence.Pat.PlatformClaims

	taID, err := arm.GetTrustAnchorID(SchemeName, token.TenantId, claims)
	if err != nil {
		return []string{""}, err
	}

	return []string{taID}, nil
}

func (s StoreHandler) GetRefValueIDs(
	tenantID string,
	trustAnchors []string,
	claims map[string]interface{},
) ([]string, error) {
	platformClaimsMap, ok := claims["cca.platform"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("claims do not contain platform map: %v", claims)
	}
	return arm.GetPlatformReferenceIDs(SchemeName, tenantID, platformClaimsMap)
}

func (s StoreHandler) SynthCoservQueryKeys(tenantID string, query string) ([]string, error) {
	return []string{"TODO"}, nil
}
