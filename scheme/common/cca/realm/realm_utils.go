// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package realm

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/veraison/ccatoken"
	"github.com/veraison/services/log"
)

func GetRIM(attr json.RawMessage) (string, error) {
	var at map[string]interface{}
	err := json.Unmarshal(attr, &at)
	if err != nil {
		return "", fmt.Errorf("unable to get RIM: %w", err)
	}
	rim, ok := at["realm-initial-measurement"].(string)
	if !ok {
		return "", errors.New("unable to get realm initial measurements")
	}
	return rim, nil
}

func GetRPV(attr json.RawMessage) ([]byte, error) {
	var at map[string]interface{}
	err := json.Unmarshal(attr, &at)
	if err != nil {
		return nil, err
	}
	iv := at["realm-personalization-value"]
	if iv == nil {
		return nil, nil
	}
	v, ok := iv.(string)
	if !ok {
		return nil, fmt.Errorf("unable to get realm personalization value from %v of type %T", iv, iv)
	}
	rpv, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return nil, err
	}
	return rpv, nil
}

func GetREMs(attr json.RawMessage) ([][]byte, error) {
	var at map[string]interface{}
	var rems [][]byte
	keys := []string{"rem0", "rem1", "rem2", "rem3"}
	err := json.Unmarshal(attr, &at)
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		rem, ok := at[key].(string)
		if ok {
			brem, err := base64.StdEncoding.DecodeString(rem)
			if err != nil {
				return nil, err
			}
			rems = append(rems, brem)
		} else {
			log.Debugf("No Rem with key = %s", key)
			break
		}
	}
	return rems, nil
}

func MapToRealmClaims(in map[string]interface{}) (ccatoken.IClaims, error) {
	realmClaims := &ccatoken.RealmClaims{}
	data, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	if err := realmClaims.FromJSON(data); err != nil {
		return nil, err
	}
	return realmClaims, nil
}

func RealmRefValLookupKey(schemeName, tenantID, rim string) string {
	absPath := []string{rim}

	u := url.URL{
		Scheme: schemeName,
		Host:   tenantID,
		Path:   strings.Join(absPath, "/"),
	}

	return u.String()
}
