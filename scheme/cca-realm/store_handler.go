// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cca_realm

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme/common"
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

	instID, err := common.GetInstID(SchemeName, refVal.Attributes)
	if err != nil {
		return nil, fmt.Errorf("unable to synthesize reference value abs-path: %w", err)
	}

	lookupKey := RefValLookupKey(SchemeName, tenantID, instID)
	log.Debugf("Scheme %s Plugin TA Look Up Key= %s\n", SchemeName, lookupKey)
	return []string{lookupKey}, nil

}

func RefValLookupKey(schemeName, tenantID, instID string) string {
	absPath := []string{instID}

	u := url.URL{
		Scheme: schemeName,
		Host:   tenantID,
		Path:   strings.Join(absPath, "/"),
	}

	return u.String()
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
