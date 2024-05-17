// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cca_realm_provisioning

import (
	"fmt"

	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme/common/arm"
)

type StoreHandler struct{}

type RealmAttr struct {
	ClassID          string    `json:"CCA_REALM.class-id"`
	Vendor           string    `json:"CCA_REALM.vendor"`
	Instance         string    `json:"CCA_REALM.inst-id"`
	HashAlgID        string    `json:"CCA_REALM.hash-alg-id"`
	MeasurementValue [5][]byte `json:"CCA_REALM.measurements"`
}

func (s StoreHandler) GetName() string {
	return "cca-realm-store-handler"
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

	lookupKey, err := arm.SynthKeyFromRefVal(SchemeName, tenantID, refVal)
	if err != nil {
		return nil, fmt.Errorf("unable to SynthKeyFromRefVal for scheme %s: %w", SchemeName, err)
	}
	log.Debugf("Scheme %s Plugin RefVal Look Up Key= %s\n", SchemeName, lookupKey)
	return []string{lookupKey}, nil
}

func (s StoreHandler) SynthKeysFromTrustAnchor(tenantID string, ta *handler.Endorsement) ([]string, error) {

	return nil, fmt.Errorf("unexpected SynthKeysFromTrustAnchor() invocation for scheme: %s", SchemeName)
}

func (s StoreHandler) GetTrustAnchorIDs(token *proto.AttestationToken) ([]string, error) {
	return nil, fmt.Errorf("unexpected GetTrustAnchorIDs() invocation for scheme: %s", SchemeName)
}

func (s StoreHandler) GetRefValueIDs(tenantID string, trustAnchors []string, claims map[string]interface{}) ([]string, error) {
	return nil, fmt.Errorf("unexpected GetRefValueIDs() invocation for scheme: %s", SchemeName)
}
