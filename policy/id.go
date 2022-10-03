// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Policy ID uniquely identifies the policy to be applied. It is currently
// assumed that, at a given point in time, there is at most one policy per
// tenant on a deployment (note: a tenant can implement multiple "logical
// policies" by defining alternate rules and switching between them in the
// top-level policy based on input). This means that a policy can effectively
// be identified (in the context of a deployment) by the tenant ID.
// Additionally, since the structure of a policy is specific to a particular
// policy agent, it is desirable to have that reflected in the ID in order to
// minimise the likelihood of issue in complex deployment that use multiple
// agents (potentially with different backends).
// With the above in mind, the policy ID is defined to be in the format
// <agent backend>://<tenant id>

// ValidateID returns nil if the provided string is a valid policy ID.
// Otherwise, an error indicating the problem is returned.
func ValidateID(in string) error {
	parts := strings.Split(in, "://")
	if len(parts) != 2 {
		return errors.New("wrong format")
	}

	if !IsValidAgentBackend(parts[0]) {
		return fmt.Errorf("not a valid agent backend: %q", parts[0])
	}

	if _, err := strconv.Atoi(parts[1]); err != nil {
		return fmt.Errorf("not a valid tenant ID: %q", parts[1])
	}

	return nil
}
