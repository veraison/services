// Copyright 2025-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package sessionmanager

import (
	"fmt"
	"net"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Memcached_SetGetDelOK(t *testing.T) {
	sm := NewMemcached()

	listner, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	server := &testServer{}
	go server.Serve(listner)

	cfg := viper.New()
	cfg.Set("servers", []string{listner.Addr().String()})

	err = sm.Init(cfg)
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
