// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_PolicyID(t *testing.T) {
	storeKey := "test-tenant:test-scheme:test-name"

	pid, err := PolicyIDFromStoreKey(storeKey)
	require.NoError(t, err)
	assert.Equal(t, "test-tenant", pid.TenantId)
	assert.Equal(t, "test-scheme", pid.Scheme)
	assert.Equal(t, "test-name", pid.Name)
	assert.Equal(t, storeKey, pid.StoreKey())

	_, err = PolicyIDFromStoreKey("bad:id")
	assert.EqualError(t, err,
		"bad policy store key \"bad:id\": want 3 :-separated parts, found 2")

	_, err = PolicyIDFromStoreKey("tenant/1:scheme:name")
	assert.EqualError(t, err,
		"bad TenantId \"tenant/1\": must be a valid URI path segment")

	_, err = PolicyIDFromStoreKey("0:bad%scheme:name")
	assert.EqualError(t, err,
		"bad Scheme \"bad%scheme\": must be a valid URI path segment")

	_, err = PolicyIDFromStoreKey("tenant:scheme:name<")
	assert.EqualError(t, err,
		"bad Name \"name<\": must be a valid URI path segment")
}

func Test_Policy_VersionedName(t *testing.T) {
	pol := Policy{
		ID:      PolicyID{"0", "PSA_IOT", "opa"},
		Rules:   "true",
		Version: 1,
	}

	assert.Equal(t, "0:opa:v1", pol.VersionedName())
}
