// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_PolicyKey(t *testing.T) {
	storeKey := "test-tenant:test-scheme:test-name"

	key, err := PolicyKeyFromString(storeKey)
	require.NoError(t, err)
	assert.Equal(t, "test-tenant", key.TenantId)
	assert.Equal(t, "test-scheme", key.Scheme)
	assert.Equal(t, "test-name", key.Name)
	assert.Equal(t, storeKey, key.String())

	_, err = PolicyKeyFromString("bad:id")
	assert.EqualError(t, err,
		"bad policy store key \"bad:id\": want 3 :-separated parts, found 2")

	_, err = PolicyKeyFromString("tenant/1:scheme:name")
	assert.EqualError(t, err,
		"bad TenantId \"tenant/1\": must be a valid URI path segment")

	_, err = PolicyKeyFromString("0:bad%scheme:name")
	assert.EqualError(t, err,
		"bad Scheme \"bad%scheme\": must be a valid URI path segment")

	_, err = PolicyKeyFromString("tenant:scheme:name<")
	assert.EqualError(t, err,
		"bad Name \"name<\": must be a valid URI path segment")
}
