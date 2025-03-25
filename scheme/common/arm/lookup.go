// Copyright 2021-2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package arm

import (
	"encoding/base64"
	"net/url"
	"strings"

	"github.com/veraison/psatoken"
)

func RefValLookupKey(schemeName, tenantID, implID string) string {
	absPath := []string{implID}

	u := url.URL{
		Scheme: schemeName,
		Host:   tenantID,
		Path:   strings.Join(absPath, "/"),
	}

	return u.String()
}

func TaLookupKey(schemeName, tenantID, implID, instID string) string {
	absPath := []string{implID, instID}

	u := url.URL{
		Scheme: schemeName,
		Host:   tenantID,
		Path:   strings.Join(absPath, "/"),
	}

	return u.String()
}

func TaCoservLookupKey(schemeName, tenantID, instID string) string {
	u := url.URL{
		Scheme: schemeName,
		Host:   tenantID,
		Path:   instID,
	}

	return u.String()
}

func MustImplIDString(c psatoken.IClaims) string {
	v, err := c.GetImplID()
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(v)
}

func MustInstIDString(c psatoken.IClaims) string {
	v, err := c.GetInstID()
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(v)
}
