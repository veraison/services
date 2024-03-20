// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cca_ssd_platform

import (
	"github.com/veraison/services/handler"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme/common/arm"
)

type StoreHandler struct{}

func (s StoreHandler) GetName() string {
	return "cca-store-handler"
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
