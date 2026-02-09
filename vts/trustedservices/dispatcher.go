// Copyright 2022-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package trustedservices

type clientDetails struct {
	Type     string
	Url      string
	Insecure bool
	CaCerts  []string
	Hints    []string
}
type DispatchInfo struct {
	ClientInfo map[string]clientDetails
}
