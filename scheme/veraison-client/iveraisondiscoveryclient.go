// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package veraisonclient

import "github.com/veraison/apiclient/verification"

// IVeraisonDiscoveryClient is an interface for dealing with Veraison's
// // apiclient/verification discovery objects
type IVeraisonDiscoveryClient interface {
	Run() (*verification.DiscoveryObject, error)
	SetDiscoveryURI(u string) error
	SetIsInsecure()
	SetCerts(paths []string) error
}
