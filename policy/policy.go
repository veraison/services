// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"errors"
	"fmt"
)

// Policy allows enforcing additional constraints on top of the regular attestation schemes.
type Policy struct {
	// ID is the identifier of this policy, unique to the store.
	ID PolicyID `json:"-"`

	// Version gets bumped every time a new policy with existing ID is added to the store.
	Version int32 `json:"version"`

	// Rules of the policy to be interpreted and execute by the policy agent.
	Rules string `json:"rules"`
}

// VersionedName returns a string identifier for the policy which is derived
// from its policy ID and version (and is therefore unique across versions).
func (o Policy) VersionedName() string {
	return fmt.Sprintf("%s:%s:v%d", o.ID.TenantId, o.ID.Name, o.Version)
}

// Validate returns an error if the cuurent policy is invalid.
func (o Policy) Validate() error {
	if o.Version == 0 {
		return errors.New("zero version")
	}

	return o.ID.Validate()
}
