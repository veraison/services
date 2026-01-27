// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package riot

import (
	"testing"

	"github.com/veraison/services/scheme/common"
)

func TestProfile(t *testing.T) {
	tcs := []common.CorimTestCase{
		{
			Title: "ok",
			Input: corimRiotValid,
		},
		{
			Title: "bad ref. vals. present",
			Input: corimRiotBadRefvals,
			Err:   "found reference values",
		},
		{
			Title: "bad no vendor",
			Input: corimRiotBadNoVendor,
			Err:   "missing vendor",
		},
		{
			Title: "bad wrong vendor",
			Input: corimRiotBadWrongVendor,
			Err:   `vendor must be "Veraison Project"`,
		},
		{
			Title: "bad wrong key type",
			Input: corimRiotBadWrongKeyType,
			Err:   `key must be a cert`,
		},
	}

	common.RunCorimTests(t, tcs)
}
