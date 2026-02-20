// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_tpm

import (
	"testing"

	"github.com/veraison/services/scheme/common"
)

func TestProfile(t *testing.T) {
	tcs := []common.CorimTestCase{
		{
			Title: "ok",
			Input: corimParsecTpmValid,
		},
		{
			Title: "bad instance",
			Input: corimParsecTpmBadInstance,
			Err:   "instance: expected bytes, found uuid",
		},
		{
			Title: "bad multiple keys",
			Input: corimParsecTpmBadMultipleKeys,
			Err:   "expected exactly one key but got 2",
		},
		{
			Title: "bad no digest",
			Input: corimParsecTpmBadNoDigests,
			Err:   "measurement 0 does not contain digests",
		},
		{
			Title: "bad no instance",
			Input: corimParsecTpmBadNoInstance,
			Err:   "instance not set in trust anchor environment",
		},
		{
			Title: "bad no class",
			Input: corimParsecTpmBadNoClass,
			Err:   "class not set",
		},
		{
			Title: "bad no class ID",
			Input: corimParsecTpmBadNoClassId,
			Err:   "class ID not set",
		},
		{
			Title: "bad class ID",
			Input: corimParsecTpmBadClassId,
			Err:   "class ID: expected uuid, found oid",
		},
		{
			Title: "bad no PCR",
			Input: corimParsecTpmBadNoPcr,
			Err:   "measurement 0 has no key",
		},
		{
			Title: "bad PCR",
			Input: corimParsecTpmBadPcr,
			Err:   "measurement 0 key: expected uint, found uuid",
		},
	}

	common.RunCorimTests(t, tcs)
}
