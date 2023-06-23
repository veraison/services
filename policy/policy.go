// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"time"

	"github.com/google/uuid"
)

// Policy allows enforcing additional constraints on top of the regular
// attestation schemes.
type Policy struct {
	// StoreKey is the identifier of this policy, unique to the store.
	StoreKey PolicyKey `json:"-"`

	// UUID is the unque identifier associated with this specific instance
	// of a policy.
	UUID uuid.UUID `json:"uuid"`

	// CTime is the creationg time of this policy.
	CTime time.Time `json:"ctime"`

	// Name is the name of this policy. It's a short descritor for the
	// rules in this policy.
	Name string `json:"name"`

	// Type identifies the policy engine used to evaluate the policy, and
	// therfore dictates how the Rules should be interpreted.
	Type string `json:"type"`

	// Rules of the policy to be interpreted and execute by the policy
	// agent.
	Rules string `json:"rules"`

	// Active indicates whether this policy instance is currently active
	// for the associated key.
	Active bool `json:"active"`
}

// NewPolicy creates a new Policy based on the specified PolicyID and rules.
func NewPolicy(key PolicyKey, name, typ, rules string) (*Policy, error) {
	polUUID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	return &Policy{
		StoreKey: key,
		UUID:     polUUID,
		CTime:    time.Now(),
		Type:     typ,
		Name:     name,
		Rules:    rules,
	}, nil
}

// Validate returns an error if the cuurent policy is invalid.
func (o *Policy) Validate() error {
	return o.StoreKey.Validate()
}
