// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_cca

import (
	"fmt"

	"github.com/veraison/services/handler"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme/common"
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

	return arm.SynthKeysFromRefValue(SchemeName, tenantID, refVal)
}

func (s StoreHandler) SynthKeysFromTrustAnchor(tenantID string, ta *handler.Endorsement) ([]string, error) {

	return arm.SynthKeysFromTrustAnchors(SchemeName, tenantID, ta)
}

func (s StoreHandler) GetTrustAnchorIDs(token *proto.AttestationToken) ([]string, error) {
	ta, err := arm.GetTrustAnchorID(SchemeName, token)
	if err != nil {
		return []string{""}, err
	}
	return []string{ta}, nil
}

func (s StoreHandler) GetRefValueIDs(
	tenantID string,
	trustAnchors []string,
	claims map[string]interface{},
) ([]string, error) {
	platformClaimsMap, ok := claims["cca.platform"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("claims to do not contain patform map: %v", claims)
	}

	platformClaims, err := common.MapToClaims(platformClaimsMap)
	if err != nil {
		return nil, err
	}

	return []string{arm.RefValLookupKey(
		SchemeName,
		tenantID,
		arm.MustImplIDString(platformClaims),
	)}, nil
}
