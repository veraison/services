// Copyright 2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package capability

import (
	"github.com/lestrrat-go/jwx/v2/jwk"
)

const (
	WellKnownMediaType = "application/vnd.veraison.discovery+json"
)

type WellKnownInfo struct {
	PublicKey                   jwk.Key           `json:"ear-verification-key,omitempty"`
	MediaTypes                  []string          `json:"media-types,omitempty"`
	CompositeEvidenceMediaTypes []string          `json:"composite-evidence-media-types,omitempty"`
	Schemes                     []string          `json:"attestation-schemes,omitempty"`
	Version                     string            `json:"version"`
	ServiceState                string            `json:"service-state"`
	ApiEndpoints                map[string]string `json:"api-endpoints"`
}

var ssTrans = map[string]string{
	"SERVICE_STATUS_UNSPECIFIED":  "UNSPECIFIED",
	"SERVICE_STATUS_DOWN":         "DOWN",
	"SERVICE_STATUS_INITIALIZING": "INITIALIZING",
	"SERVICE_STATUS_READY":        "READY",
	"SERVICE_STATUS_TERMINATING":  "TERMINATING",
}

func ServiceStateToAPI(ss string) string {
	t, ok := ssTrans[ss]
	if !ok {
		return "UNKNOWN"
	}
	return t
}

func NewWellKnownInfoObj(
	key jwk.Key,
	mediaTypes []string,
	compositeEvidenceMediaTypes []string,
	schemes []string,
	version string,
	serviceState string,
	endpoints map[string]string,
) (*WellKnownInfo, error) {
	// MUST be kept in sync with proto/state.proto
	obj := &WellKnownInfo{
		PublicKey:                   key,
		MediaTypes:                  mediaTypes,
		CompositeEvidenceMediaTypes: compositeEvidenceMediaTypes,
		Schemes:                     schemes,
		Version:                     version,
		ServiceState:                ServiceStateToAPI(serviceState),
		ApiEndpoints:                endpoints,
	}

	return obj, nil
}
