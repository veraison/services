// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package sevsnp

import (
	_ "embed"
	"testing"

	"github.com/veraison/services/pkg/util"
)

func TestProfile(t *testing.T) {
	tcs := []util.CorimTestCase{
		{
			Title: "ok",
			Input: corimSevsnpValid,
		},
		{
			Title: "bad TA env no model",
			Input: corimSevsnpBadTaNoModel,
			Err:   "missing model",
		},
		{
			Title: "bad TA env no vendor",
			Input: corimSevsnpBadTaNoVendor,
			Err:   "missing vendor",
		},
		{
			Title: "bad RefVal no mkey",
			Input: corimSevsnpBadRefvalNoKey,
			Err:   "mkey not set",
		},
		{
			Title: "bad RefVal mkey type",
			Input: corimSevsnpBadRefvalKey,
			Err:   "mkey type: expected uint, got oid",
		},
	}

	util.RunCorimTests(t, tcs)
}
