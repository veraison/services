// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"fmt"

	"github.com/veraison/services/config"
)

var DefaultBackend = "opa"

var ErrBadResult = "could not create updated AttestationResult: %w from JSON %s"
var ErrNoStatus = "backend returned outcome with no status field: %v"
var ErrNoTV = "backend returned no trust-vector field, or its not a map[string]interface{}: %v"

// CreateAgent creates a new PolicyAgent using the backend specified in the
// config with "policy.backend" directive. If this directive is absent, the
// default backend, "opa",  will be used.
func CreateAgent(cfg config.Store) (IAgent, error) {
	backendString, err := config.GetString(cfg, DirectiveBackend, &DefaultBackend)
	if err != nil {
		return nil, fmt.Errorf("loading backend from config: %w", err)
	}

	var backend IBackend

	switch backendString {
	case "opa":
		backend = &OPA{}
	default:
		return nil, fmt.Errorf("backend %q is not supported", backendString)
	}

	return &PolicyAgent{Backend: backend}, nil
}
