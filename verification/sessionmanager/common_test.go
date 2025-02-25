// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package sessionmanager

import (
	"time"

	"github.com/google/uuid"
)

var (
	testTenant      = "0123456789"
	testUUIDString  = uuid.NewString()
	testUUID        = uuid.MustParse(testUUIDString)
	testSession     = []byte(`{ "a": 1 }`)
	testTTL, _      = time.ParseDuration("1m30s")
	testShortTTL, _ = time.ParseDuration("1s")
)

