// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package arm_cca

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/veraison/ccatoken"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/coserv"
	"github.com/veraison/services/handler"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/scheme/common/arm"
	"github.com/veraison/services/scheme/common/cca/realm"
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

func (s StoreHandler) SynthCoservQueryKeys(tenantID string, query string) ([]string, error) {
	var q coserv.Coserv
	if err := q.FromBase64Url(query); err != nil {
		return nil, err
	}

	var keys []string

	switch q.Query.ArtifactType {
	case coserv.ArtifactTypeReferenceValues:
		s := q.Query.EnvironmentSelector

		if s.Classes != nil {
			for i, v := range *s.Classes {
				implID, err := extractImplID(*v.Class)
				if err != nil {
					return nil, fmt.Errorf("creating lookup key for class[%d]: %w", i, err)
				}

				keys = append(keys, arm.RefValLookupKey(SchemeName, tenantID, implID))
			}
		}
	case coserv.ArtifactTypeTrustAnchors:
		s := q.Query.EnvironmentSelector

		if s.Instances != nil {
			for i, v := range *s.Instances {
				instID, err := extractInstID(*v.Instance)
				if err != nil {
					return nil, fmt.Errorf("creating lookup key for instance[%d]: %w", i, err)
				}

				keys = append(keys, arm.TaCoservLookupKey(SchemeName, tenantID, instID))
			}
		}
	case coserv.ArtifactTypeEndorsedValues:
		return nil, errors.New("CCA does not implement endorsed value queries")
	}

	return keys, nil
}

func extractImplID(c comid.Class) (string, error) {
	if c.ClassID == nil {
		return "", errors.New("missing class-id")
	}

	implID, err := c.ClassID.GetImplID()
	if err != nil {
		return "", fmt.Errorf("could not extract implementation-id from class-id: %w", err)
	}

	return implID.String(), nil
}

func extractInstID(i comid.Instance) (string, error) {
	instID, err := i.GetUEID()
	if err != nil {
		return "", fmt.Errorf("could not extract implementation-id from instance-id: %w", err)
	}

	return base64.StdEncoding.EncodeToString(instID), nil
}
