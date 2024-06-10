// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package cca

import (
	"encoding/base64"
	"fmt"

	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
)

func SynthKeysForCcaRealm(scheme string, tenantID string,
	refVal *handler.Endorsement,
) ([]string, error) {
	rim, err := GetRim(refVal.SubScheme, refVal.Attributes)
	if err != nil {
		return nil, fmt.Errorf("unable to get rim %w", err)
	}
	lookupKey := RealmRefValLookupKey(scheme, tenantID, rim)
	log.Debugf("Scheme %s Plugin Reference Value Look Up Key= %s\n", scheme, lookupKey)
	return []string{lookupKey}, nil
}

func GetRealmReferenceIDs(
	scheme string,
	tenantID string,
	claims map[string]interface{},
) ([]string, error) {
	realmClaimsMap, ok := claims["cca.realm"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("claims do not contain realm map: %v", claims)
	}

	realmClaims, err := MapToRealmClaims(realmClaimsMap)
	if err != nil {
		return nil, err
	}
	m, err := realmClaims.GetInitialMeasurement()
	if err != nil {
		return nil, err
	}
	rim := base64.StdEncoding.EncodeToString(m)
	return []string{RealmRefValLookupKey(scheme, tenantID, rim)}, nil
}
