// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

// DefaultBackend will be used if backend is not explicitly specfied
var DefaultBackend = "opa"

var backends = map[string]IBackend{
	"opa": &OPA{},
}

// IsValidAgentBackend returns True iff the specified string names a valid backend.
func IsValidAgentBackend(name string) bool {
	_, ok := backends[name]
	return ok
}

// GetSupportedBackends returns a string slice of supported backend names.
func GetSupportedAgentBackends() []string {
	var names []string // nolint:prealloc

	for name := range backends {
		names = append(names, name)
	}

	return names
}
