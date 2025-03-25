// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_tpm

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/veraison/parsec/tpm"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
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

func (s StoreHandler) GetRefValueIDs(
	tenantID string,
	trustAnchors []string,
	claims map[string]interface{},
) ([]string, error) {
	var endorsement TaEndorsements

	if err := json.Unmarshal([]byte(trustAnchors[0]), &endorsement); err != nil {
		log.Errorf("Could not decode Endorsements in ExtractClaims: %v", err)
		return nil, fmt.Errorf("could not decode endorsement: %w", err)
	}

	class_id := *endorsement.Attr.ClassID
	return []string{tpmLookupKey(ScopeRefValues, tenantID, class_id, "")}, nil
}

func synthKeysFromAttr(scope, tenantID string, attr json.RawMessage) ([]string, error) {
	var (
		instance string
		class    string
		err      error
	)

	switch scope {
	case ScopeTrustAnchor:
		var ta TaAttr
		if err := json.Unmarshal(attr, &ta); err != nil {
			return nil, fmt.Errorf("unable to extract endorsements from TA: %w", err)
		}
		if ta.ClassID == nil || ta.InstID == nil {
			return nil, fmt.Errorf("missing InstID or ClassID from TA: %w", err)
		}
		class = *ta.ClassID
		instance = *ta.InstID
	case ScopeRefValues:
		var sw SwAttr
		if err := json.Unmarshal(attr, &sw); err != nil {
			return nil, fmt.Errorf("unable to extract endorsements from RefVal: %w", err)
		}
		if sw.ClassID == nil {
			return nil, fmt.Errorf("missing ClassID in reference value: %w", err)
		}
		class = *sw.ClassID
	default:
		return nil, fmt.Errorf("invalid scope argument: %s", scope)
	}

	return []string{tpmLookupKey(scope, tenantID, class, instance)}, nil
}

func tpmLookupKey(scope, tenantID, class, instance string) string {
	var absPath []string

	switch scope {
	case ScopeTrustAnchor:
		absPath = []string{instance}
	case ScopeRefValues:
		absPath = []string{class}
	}

	u := url.URL{
		Scheme: SchemeName,
		Host:   tenantID,
		Path:   strings.Join(absPath, "/"),
	}

	return u.String()
}

func (s StoreHandler) SynthCoservQueryKeys(tenantID string, query string) ([]string, error) {
	return []string{"TODO"}, nil
}
