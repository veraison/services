// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package trustedservices

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewDispatcher_ok(t *testing.T) {
	fp := "testdata/dispatch-table.json"
	_, err := NewDispatcher(fp)
	require.NoError(t, err)

}
