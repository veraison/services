// Copyright 2021-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package psa_iot

import (
	"github.com/veraison/psatoken"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme/common/arm"
)

type StoreHandler struct{}

func (s StoreHandler) GetName() string {
	return "psa-store-handler"
}

func (s StoreHandler) GetAttestationScheme() string {
	return SchemeName
}

func (s StoreHandler) GetSupportedMediaTypes() []string {
	return nil
}

func (s StoreHandler) SynthKeysFromRefValue(
	tenantID string,
	refValue *handler.Endorsement,
) ([]string, error) {
	return arm.SynthKeysForPlatform(SchemeName, tenantID, refValue)
}

func (s StoreHandler) SynthKeysFromTrustAnchor(tenantID string, ta *handler.Endorsement) ([]string, error) {
	return arm.SynthKeysFromTrustAnchors(SchemeName, tenantID, ta)
}

func (s StoreHandler) GetTrustAnchorIDs(token *proto.AttestationToken) ([]string, error) {
	psaToken, err := psatoken.DecodeAndValidateEvidenceFromCOSE(token.Data)
	if err != nil {
		return []string{""}, handler.BadEvidence(err)
	}

	claims := psaToken.Claims

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
	return arm.GetPlatformReferenceIDs(SchemeName, tenantID, claims)
}

func (s StoreHandler) SynthCoservQueryKeys(tenantID string, query string) ([]string, error) {
	return []string{"TODO"}, nil
}
