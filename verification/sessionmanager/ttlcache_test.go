// Copyright 2022-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package sessionmanager

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_TTLCache_SetGetDelOK(t *testing.T) {
	sm := TTLCache{}

	err := sm.Init(nil)
	defer sm.Close()

	assert.NoError(t, err)

	err = sm.SetSession(testUUID, testTenant, testSession, testTTL)
	assert.NoError(t, err)

	session, err := sm.GetSession(testUUID, testTenant)
	assert.NoError(t, err)
	assert.JSONEq(t, string(testSession), string(session))

	err = sm.DelSession(testUUID, testTenant)
	assert.NoError(t, err)

	expectedErr := fmt.Sprintf("session not found for (id, tenant)=(%s, %s)", testUUIDString, testTenant)

	_, err = sm.GetSession(testUUID, testTenant)
	assert.EqualError(t, err, expectedErr)
}

func Test_TTLCache_Eviction(t *testing.T) {
	sm := TTLCache{}

	err := sm.Init(nil)
	defer sm.Close()

	assert.NoError(t, err)

	err = sm.SetSession(testUUID, testTenant, testSession, testShortTTL)
	assert.NoError(t, err)

	// wait enough for eviction to kick in
	time.Sleep(2 * testShortTTL)

	expectedErr := fmt.Sprintf("session not found for (id, tenant)=(%s, %s)", testUUIDString, testTenant)

	// check that the previously Set session is gone
	_, err = sm.GetSession(testUUID, testTenant)
	assert.EqualError(t, err, expectedErr)
}
