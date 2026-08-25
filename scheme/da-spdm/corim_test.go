// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package da_spdm

import (
	"testing"

	"github.com/veraison/services/scheme/common"
)

func TestProfile(t *testing.T) {
	tcs := []common.CorimTestCase{
		{
			Title: "ok",
			Input: corimDaRef,
		},
		{
			Title: "bad no class",
			Input: corimDaBadNoClass,
			Err: "class not set",
		},
		{
			Title: "bad no class ID",
			Input: corimDaBadNoClassId,
			Err: "class ID not set",
		},
		{
			Title: "bad instance set",
			Input: corimDaBadInstanceSet,
			Err: "instance must not be set",
		},
		{
			Title: "bad group set",
			Input: corimDaBadGroupSet,
			Err: "group must not be set",
		},
		{
			Title: "bad mkey not set",
			Input: corimDaBadMkeyNotSet,
			Err: "measurement[0]: mkey not set",
		},
		{
			Title: "bad mkey wrong type",
			Input: corimDaBadMkeyWrongType,
			Err: "measurement[0]: measurement-key type is: *comid.StringMkey",
		},
		{
			Title: "bad no measurement name",
			Input: corimDaBadNoName,
			Err: "measurement[0]: name not set",
		},
		{
			Title: "bad no measurement value",
			Input: corimDaBadNoValue,
			Err: "measurement[0]: either digests or raw-value must be set",
		},
		{
			Title: "bad both digests and raw-value are set",
			Input: corimDaBadMultipleValues,
			Err: "measurement[0]: digests and raw-value can't both be set",
		},
	}

	common.RunCorimTests(t, tcs)
}
