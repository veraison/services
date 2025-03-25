// Copyright 2021-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package riot

import (
	"errors"

	"github.com/veraison/services/handler"
	"github.com/veraison/services/proto"
)

type StoreHandler struct {
}

func (s StoreHandler) GetName() string {
	return "riot-store-handler"
}

func (s StoreHandler) GetAttestationScheme() string {
	return SchemeName
}

func (s StoreHandler) GetSupportedMediaTypes() []string {
	return nil
}

func (s StoreHandler) GetTrustAnchorIDs(token *proto.AttestationToken) ([]string, error) {
	return []string{"dice://"}, nil
}

func (s StoreHandler) GetRefValueIDs(
	tenantID string,
	trustAnchors []string,
	claims map[string]interface{},
) ([]string, error) {
	return []string{"dice://"}, nil
}

func (s StoreHandler) SynthKeysFromRefValue(tenantID string, swComp *handler.Endorsement) ([]string, error) {
	return nil, errors.New("TODO")
}

func (s StoreHandler) SynthKeysFromTrustAnchor(tenantID string, ta *handler.Endorsement) ([]string, error) {
	return nil, errors.New("TODO")
}

func (s StoreHandler) SynthCoservQueryKeys(tenantID string, query string) ([]string, error) {
	return []string{"TODO"}, nil
}
