// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_tpm

import (
	"encoding/base64"
	"errors"

	"github.com/veraison/parsec/tpm"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/proto"
)

type StoreHandler struct{}

func (s StoreHandler) GetName() string {
	return "parsec-tpm-store-handler"
}

func (s StoreHandler) GetAttestationScheme() string {
	return SchemeName
}

func (s StoreHandler) GetSupportedMediaTypes() []string {
	return nil
}

func (s StoreHandler) SynthKeysFromRefValue(tenantID string, refVals *handler.Endorsement) ([]string, error) {
	return synthKeysFromAttr(ScopeRefValues, tenantID, refVals.Attributes)
}

func (s StoreHandler) SynthKeysFromTrustAnchor(tenantID string, ta *handler.Endorsement) ([]string, error) {
	return synthKeysFromAttr(ScopeTrustAnchor, tenantID, ta.Attributes)
}

func (s StoreHandler) GetTrustAnchorIDs(token *proto.AttestationToken) ([]string, error) {
	var ev tpm.Evidence
	err := ev.FromCBOR(token.Data)
	if err != nil {
		return []string{""}, handler.BadEvidence(err)
	}

	kat := ev.Kat
	if kat == nil {
		return []string{""}, errors.New("no key attestation token to fetch Key ID")
	}
	kid := *kat.KID
	instance_id := base64.StdEncoding.EncodeToString(kid)
	return []string{tpmLookupKey(ScopeTrustAnchor, token.TenantId, "", instance_id)}, nil

}
