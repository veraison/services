// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"fmt"
	"net/url"
	"strings"
)

// PolicyID identifies a specific policy. This is used to retrieve the policy
// from the store.
type PolicyID struct {
	// TenantId  is the ID of the tenant that owns this policy.
	TenantId string `json:"tenant_id"`

	// Scheme is the name of the scheme with which this policy is associated
	Scheme string `json:"scheme"`

	// Name is the name of this policy
	Name string `json:"name"`
}

// PolicyIDFromKey parses the specified string containing a policy store key
// into a PolicyID.
func PolicyIDFromStoreKey(s string) (PolicyID, error) {
	var pid PolicyID

	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return pid, fmt.Errorf(
			"bad policy store key %q: want 3 :-separated parts, found %d",
			s, len(parts),
		)
	}

	pid.TenantId = parts[0]
	pid.Scheme = parts[1]
	pid.Name = parts[2]

	return pid, pid.Validate()
}

func (o PolicyID) Validate() error {
	if url.PathEscape(o.TenantId) != o.TenantId {
		return fmt.Errorf("bad TenantId %q: must be a valid URI path segment", o.TenantId)
	}

	if url.PathEscape(o.Scheme) != o.Scheme {
		return fmt.Errorf("bad Scheme %q: must be a valid URI path segment", o.Scheme)
	}

	if url.PathEscape(o.Name) != o.Name {
		return fmt.Errorf("bad Name %q: must be a valid URI path segment", o.Name)
	}

	return nil
}

// StoreKey returns the key used with the policy store.
func (o PolicyID) StoreKey() string {
	return fmt.Sprintf("%s:%s:%s", o.TenantId, o.Scheme, o.Name)
}
