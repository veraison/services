// Copyright 2022-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/services/log"
)

func Test_Store_CRUD(t *testing.T) {
	v := viper.New()
	v.Set("backend", "memory")

	store, err := NewStore(v, log.Named("test"))
	require.NoError(t, err)
	defer store.Close()

	key := PolicyKey{"1", "scheme", "policy"}

	policy, err := store.Add(key, "test", "test",
		"1. the chief's always right; 2. if the chief's wrong, see 1.")
	require.NoError(t, err)

	_, err = store.GetActive(key)
	assert.EqualError(t, err, "no active policy for key \"1:scheme:policy\"")

	err = store.Activate(key, policy.UUID)
	require.NoError(t, err)

	policy, err = store.GetActive(key)
	require.NoError(t, err)
	assert.Equal(t, key, policy.StoreKey)

	_, err = store.Add(key, "test", "test", "On second thought, chief's not always right.")
	assert.ErrorContains(t, err, "already exists")

	secondPolicy, err := store.Update(key, "test", "test",
		"On second thought, chief's not always right.")
	require.NoError(t, err)

	policy, err = store.GetActive(key)
	require.NoError(t, err)
	assert.Equal(t, "1. the chief's always right; 2. if the chief's wrong, see 1.", policy.Rules)

	err = store.Activate(key, secondPolicy.UUID)
	require.NoError(t, err)

	policy, err = store.GetActive(key)
	require.NoError(t, err)
	assert.Equal(t, "On second thought, chief's not always right.", policy.Rules)

	err = store.DeactivateAll(key)
	require.NoError(t, err)

	_, err = store.GetActive(key)
	assert.EqualError(t, err, "no active policy for key \"1:scheme:policy\"")

	versions, err := store.Get(key)
	require.NoError(t, err)
	assert.Len(t, versions, 2)

	policies, err := store.List()
	require.NoError(t, err)
	assert.Equal(t, 0, len(policies))

	err = store.Activate(key, secondPolicy.UUID)
	require.NoError(t, err)

	policies, err = store.List()
	require.NoError(t, err)
	assert.Equal(t, 1, len(policies))
	assert.Equal(t, key, policies[0].StoreKey)

	policies, err = store.ListAllVersions()
	require.NoError(t, err)
	assert.Equal(t, 2, len(policies))
	assert.Equal(t, key, policies[0].StoreKey)
	assert.NotEqual(t, policies[0].UUID, policies[1].UUID)

	err = store.Del(key)
	require.NoError(t, err)

	_, err = store.GetActive(key)
	assert.ErrorIs(t, err, ErrNoPolicy)
}
