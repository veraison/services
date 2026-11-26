// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package vts

type State struct {
	Version                string
	ProvisioningMediaTypes []string
	VerificationMediaTypes []string
	CoservMediaTypes       []string
	CorimProfiles          []string
	EarSigningPublicKey    string
	CoservSigningPublicKey string
}
