// Copyright 2021-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/services/log"
)

var (
	testKey    = `psa://tenant-1/deadbeef/beefdead`
	testVal    = `{"some": "json"}`
	altTestKey = `psa://tenant-2/cafecafe/cafecafe`
	altTestVal = `[1, 2, 3]`
)

func TestMemory_Init_Close_cycle_ok(t *testing.T) {
	s := Memory{}

	err := s.Init(nil, log.Named("test"))
	assert.NoError(t, err)
	assert.Len(t, s.Data, 0)

	err = s.Close()
	assert.NoError(t, err)
}

func TestMemory_Set_Get_Del_with_uninitialised_store(t *testing.T) {
	s := Memory{}

	expectedErr := `memory store uninitialized`

	err := s.Set(testKey, testVal)
	assert.EqualError(t, err, expectedErr)

	err = s.Del(testKey)
	assert.EqualError(t, err, expectedErr)

	_, err = s.Get(testKey)
	assert.EqualError(t, err, expectedErr)
}

func TestMemory_Set_Get_ok(t *testing.T) {
	s := Memory{}

	err := s.Init(nil, log.Named("test"))
	require.NoError(t, err)

	err = s.Set(testKey, testVal)
	assert.NoError(t, err)

	expectedVal := []string{testVal}

	val, err := s.Get(testKey)
	assert.NoError(t, err)
	assert.Equal(t, expectedVal, val)

	keys, err := s.GetKeys()
	assert.NoError(t, err)
	require.Equal(t, 1, len(keys))
	assert.Equal(t, testKey, keys[0])
}

func TestMemory_GetMultiple(t *testing.T) {
	s := Memory{}

	err := s.Init(nil, log.Named("test"))
	require.NoError(t, err)

	s.Data = map[string][]string{
		"key1": []string{"1", "2"},
		"key2": []string{"3", "4"},
	}

	ret, err := s.GetMultiple([]string{"key1", "key2"})
	require.NoError(t, err)
	assert.Equal(t, []string{"1", "2", "3", "4"}, ret)

	ret, err = s.GetMultiple([]string{"key2", "key1"})
	require.NoError(t, err)
	assert.Equal(t, []string{"3", "4", "1", "2"}, ret)

	ret, err = s.GetMultiple([]string{"key1", "key2", "key3"})
	assert.EqualError(t, err, `key not found: "key3"`)
	assert.Equal(t, []string{"1", "2", "3", "4"}, ret)
}

func TestMemory_Get_empty_key(t *testing.T) {
	s := Memory{}

	err := s.Init(nil, log.Named("test"))
	require.NoError(t, err)

	emptyKey := ""
	expectedErr := `the supplied key is empty`

	_, err = s.Get(emptyKey)
	assert.EqualError(t, err, expectedErr)
}

func TestMemory_Del_empty_key(t *testing.T) {
	s := Memory{}

	err := s.Init(nil, log.Named("test"))
	require.NoError(t, err)

	emptyKey := ""
	expectedErr := `the supplied key is empty`

	err = s.Del(emptyKey)
	assert.EqualError(t, err, expectedErr)
}

func TestMemory_Set_empty_key(t *testing.T) {
	s := Memory{}

	err := s.Init(nil, log.Named("test"))
	require.NoError(t, err)

	emptyKey := ""
	expectedErr := `the supplied key is empty`

	err = s.Set(emptyKey, testVal)
	assert.EqualError(t, err, expectedErr)
}

func TestMemory_Set_bad_json(t *testing.T) {
	s := Memory{}

	err := s.Init(nil, log.Named("test"))
	require.NoError(t, err)

	badJSON := "[1, 2"
	expectedErr := `the supplied val contains invalid JSON: unexpected end of JSON input`

	err = s.Set(testKey, badJSON)
	assert.EqualError(t, err, expectedErr)
}

func TestMemory_Add_using_same_key_different_vals(t *testing.T) {
	s := Memory{}

	err := s.Init(nil, log.Named("test"))
	require.NoError(t, err)

	err = s.Set(testKey, testVal)
	require.NoError(t, err)

	err = s.Add(testKey, altTestVal)
	assert.NoError(t, err)

	expectedVal := []string{testVal, altTestVal}

	val, err := s.Get(testKey)
	assert.NoError(t, err)
	assert.Equal(t, expectedVal, val)
}

func TestMemory_Add_using_same_key_same_vals(t *testing.T) {
	s := Memory{}

	err := s.Init(nil, log.Named("test"))
	require.NoError(t, err)

	err = s.Set(testKey, testVal)
	require.NoError(t, err)

	err = s.Add(testKey, testVal)
	assert.NoError(t, err)

	expectedVal := []string{testVal}

	val, err := s.Get(testKey)
	assert.NoError(t, err)
	assert.Equal(t, expectedVal, val)
}

func TestMemory_Del_ok(t *testing.T) {
	s := Memory{}

	err := s.Init(nil, log.Named("test"))
	require.NoError(t, err)

	err = s.Set(testKey, testVal)
	require.NoError(t, err)

	expectedVal := []string{testVal}

	val, err := s.Get(testKey)
	assert.NoError(t, err)
	assert.Equal(t, expectedVal, val)

	err = s.Del(testKey)
	assert.NoError(t, err)

	expectedErr := fmt.Sprintf("key not found: %q", testKey)

	_, err = s.Get(testKey)
	assert.EqualError(t, err, expectedErr)
}

func TestMemory_Get_no_such_key(t *testing.T) {
	s := Memory{}

	err := s.Init(nil, log.Named("test"))
	require.NoError(t, err)

	expectedErr := fmt.Sprintf("key not found: %q", testKey)

	_, err = s.Get(testKey)
	assert.ErrorIs(t, err, ErrKeyNotFound)
	assert.EqualError(t, err, expectedErr)
}

func TestMemory_Del_no_such_key(t *testing.T) {
	s := Memory{}

	err := s.Init(nil, log.Named("test"))
	require.NoError(t, err)

	expectedErr := fmt.Sprintf("key not found: %q", testKey)

	err = s.Del(testKey)
	assert.ErrorIs(t, err, ErrKeyNotFound)
	assert.EqualError(t, err, expectedErr)
}

func TestMemory_dump_ok(t *testing.T) {
	s := Memory{}

	err := s.Init(nil, log.Named("test"))
	require.NoError(t, err)

	err = s.Set(testKey, testVal)
	require.NoError(t, err)
	err = s.Set(altTestKey, altTestVal)
	require.NoError(t, err)

	expectedTbl := `Key                              Val
---                              ---
psa://tenant-1/deadbeef/beefdead [{"some": "json"}]
psa://tenant-2/cafecafe/cafecafe [[1, 2, 3]]
`
	tbl := s.dump()
	assert.Equal(t, expectedTbl, tbl)

}
