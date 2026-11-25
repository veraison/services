// Copyright 2023-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"github.com/lestrrat-go/jwx/v2/jwk"
)

const (
	WellKnownMediaType = "application/vnd.veraison.discovery+json"
)

type WellKnownInfo struct {
	PublicKey    jwk.Key           `json:"ear-verification-key,omitempty"`
	MediaTypes   []string          `json:"media-types,omitempty"`
	Schemes      []string          `json:"attestation-schemes,omitempty"`
	Version      string            `json:"version"`
	ApiEndpoints map[string]string `json:"api-endpoints"`
}

func NewWellKnownInfoObj(
	key jwk.Key,
	mediaTypes []string,
	schemes []string,
	version string,
	endpoints map[string]string,
) (*WellKnownInfo, error) {
	// MUST be kept in sync with proto/state.proto
	obj := &WellKnownInfo{
		PublicKey:    key,
		MediaTypes:   mediaTypes,
		Schemes:      schemes,
		Version:      version,
		ApiEndpoints: endpoints,
	}

	return obj, nil
}
