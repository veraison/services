// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package arm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/veraison/services/handler"
	"github.com/veraison/services/log"
)

func GetRim(scheme string, attr json.RawMessage) (string, error) {
	var at map[string]interface{}
	err := json.Unmarshal(attr, &at)
	if err != nil {
		return "", fmt.Errorf("unable to get Instance ID: %w", err)
	}
	key := scheme + ".realm-initial-measurement"
	rim, ok := at[key].(string)
	if !ok {
		return "", errors.New("unable to get realm initial measurements")
	}
	return rim, nil
}

func GetRpv(scheme string, attr json.RawMessage) (string, error) {
	var at map[string]interface{}
	err := json.Unmarshal(attr, &at)
	if err != nil {
		return "", fmt.Errorf("unable to get Instance ID: %w", err)
	}
	key := scheme + ".realm-personalization-value"
	rpv, ok := at[key].(string)
	if !ok {
		return "", errors.New("unable to get realm personalization value")
	}
	return rpv, nil
}

func SynthKeyFromRefVal(scheme string, tenantID string, refVal *handler.Endorsement) (string, error) {
	if refVal == nil {
		return "", errors.New("no reference value in SynthKeyFromRefVal")
	}
	rim, err := GetRim(scheme, refVal.Attributes)
	if err != nil {
		return "", fmt.Errorf("unable to get rim: %w", err)
	}
	rpv, err := GetRpv(scheme, refVal.Attributes)
	if err != nil {
		return "", fmt.Errorf("unable to get rpv: %w", err)
	}
	lookupKey := refValLookupKey(scheme, tenantID, rim, rpv)
	log.Debugf("Scheme %s realm RefVal Look Up Key= %s\n", scheme, lookupKey)
	return lookupKey, nil
}

func refValLookupKey(schemeName, tenantID, rim string, rpv string) string {
	var absPath []string
	if rpv != "" {
		absPath = []string{rim, rpv}
	} else {
		absPath = []string{rim}
	}

	u := url.URL{
		Scheme: schemeName,
		Host:   tenantID,
		Path:   strings.Join(absPath, "/"),
	}
	return u.String()
}
