// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package api

import (
	"github.com/lestrrat-go/jwx/v2/jwk"
)

const (
	CoservDiscoveryMediaType = "application/coserv-discovery+json"
)

type Capability struct {
	MediaType       string          `json:"media-type"`
	ArtifactSupport ArtifactSupport `json:"artifact-support"`
}

type ArtifactSupport []string

type CoservWellKnownInfo struct {
	Version                string            `json:"version,omitempty"`
	Capabilities           []Capability      `json:"capabilities,omitempty"`
	ApiEndpoints           map[string]string `json:"api-endpoints,omitempty"`
	ResultVerificationKeys []jwk.Key         `json:"result-verification-key,omitempty"`
}

func NewCoservWellKnownInfo(
	version string,
	capabilities []Capability,
	apiEndpoints map[string]string,
	resultVerificationKeys []jwk.Key,
) *CoservWellKnownInfo {
	return &CoservWellKnownInfo{
		Version:                version,
		Capabilities:           capabilities,
		ResultVerificationKeys: resultVerificationKeys,
		ApiEndpoints:           apiEndpoints,
	}
}
