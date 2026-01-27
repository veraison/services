// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package tpm_enacttrust

import (
	"testing"

	"github.com/veraison/services/scheme/common"
)

func TestProfile(t *testing.T) {
	tcs := []common.CorimTestCase{
		{
			Title: "ok",
			Input: corimEnacttrustValid,
		},
		{
			Title: "bad class",
			Input: corimEnacttrustBadClass,
			Err:   "class set in environment",
		},
		{
			Title: "bad instance",
			Input: corimEnacttrustBadInstance,
			Err:   "instance: expected uuid, found ueid",
		},
		{
			Title: "bad multiple keys",
			Input: corimEnacttrustBadMultipleKeys,
			Err:   "expected trust anchor to contain exactly one key; found 2",
		},
		{
			Title: "bad no digest",
			Input: corimEnacttrustBadNoDigest,
			Err:   "no digests in measurement",
		},
		{
			Title: "bad no instance",
			Input: corimEnacttrustBadNoInstance,
			Err:   "instance not set in environment",
		},
		{
			Title: "bad multiple measurements",
			Input: corimEnacttrustBadMultipleMeasurements,
			Err:   "expected exactly one measurement, found 2",
		},
		{
			Title: "bad multiple digests",
			Input: corimEnacttrustBadMultipleDigests,
			Err:   "expected exactly one digest in measurement, found 2",
		},
	}

	common.RunCorimTests(t, tcs)
}
