// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package arm_cca

import (
	"fmt"

	"github.com/veraison/ccatoken"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme/common/arm"
	"github.com/veraison/services/scheme/common/cca/realm"
)

type StoreHandler struct{}

func (s StoreHandler) GetName() string {
	return StoreHandlerName
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
	switch refVal.SubType {
	case "platform.sw-component", "platform.config":
		return arm.SynthKeysForPlatform(SchemeName, tenantID, refVal)
	case "realm.reference-value":
		return realm.SynthKeysForCcaRealm(SchemeName, tenantID, refVal)
	default:
		return nil, fmt.Errorf("invalid SubType: %s, for Scheme: %s", refVal.SubType, refVal.Scheme)
	}
}

func (s StoreHandler) SynthKeysFromTrustAnchor(tenantID string, ta *handler.Endorsement) ([]string, error) {
	return arm.SynthKeysFromTrustAnchors(SchemeName, tenantID, ta)
}

func (s StoreHandler) GetTrustAnchorIDs(token *proto.AttestationToken) ([]string, error) {
	evidence, err := ccatoken.DecodeAndValidateEvidenceFromCBOR(token.Data)
	if err != nil {
		return []string{""}, handler.BadEvidence(err)
	}

	claims := evidence.PlatformClaims
	if err != nil {
		return []string{""}, err
	}
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
	platformClaimsMap, ok := claims["platform"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("claims do not contain platform map: %v", claims)
	}
	pids, err := arm.GetPlatformReferenceIDs(SchemeName, tenantID, platformClaimsMap)
	if err != nil {
		return nil, fmt.Errorf("unable to get cca platform reference IDs: %w", err)
	}
	realmClaimsMap, ok := claims["realm"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("claims do not contain realm map: %v", claims)
	}
	rids, err := realm.GetRealmReferenceIDs(SchemeName, tenantID, realmClaimsMap)
	if err != nil {
		return nil, fmt.Errorf("unable to get cca realm reference IDs: %w", err)
	}
	return append(pids, rids...), nil
}
