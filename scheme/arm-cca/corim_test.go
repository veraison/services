// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package arm_cca

import (
	"testing"

	"github.com/veraison/services/scheme/common"
)

func TestProfile(t *testing.T) {
	tcs := []common.CorimTestCase{
		{
			Title: "platform ok",
			Input: corimCcaPlatformValid,
		},
		{
			Title: "platform bad no class",
			Input: corimCcaPlatformBadNoClass,
			Err: "class not set",
		},
		{
			Title: "platform bad TA no instance",
			Input: corimCcaPlatformBadTaNoInstance,
			Err: "instance not set for trust anchor",
		},
		{
			Title: "platform bad TA bytes instance",
			Input: corimCcaPlatformBadTaInstance,
			Err: "instance: expected UEID, got bytes",
		},
		{
			Title: "platform bad TA cert",
			Input: corimCcaPlatformBadTaCert,
			Err: "trust anchor must be a PKIX base64 key, found: pkix-base64-cert",
		},
		{
			Title: "platform bad RefVal instance",
			Input: corimCcaPlatformBadRefvalInstance,
			Err: "instance set for reference value",
		},
		{
			Title: "platform bad RefVal no mkey",
			Input: corimCcaPlatformBadRefvalNoMkey,
			Err: "measurement 0 key not set",
		},
		{
			Title: "platform bad RefVal uint mkey",
			Input: corimCcaPlatformBadRefvalMkey,
			Err: "measurement 0 key: unexpected type uint",
		},
		{
			Title: "platform bad RefVal no digest",
			Input: corimCcaPlatformBadRefvalNoDigests,
			Err: "measurement 0 value: no digests",
		},
		{
			Title: "platform bad RefVal no raw value",
			Input: corimCcaPlatformBadRefvalNoRawValue,
			Err: "measurement 0 value: no raw value",
		},
		{
			Title: "realm ok",
			Input: corimCcaRealmValid,
		},
		{
			Title: "realm bad instance",
			Input: corimCcaRealmBadInstance,
			Err: "instance: expected bytes, got ueid",
		},
		{
			Title: "realm bad no instance",
			Input: corimCcaRealmBadNoInstance,
			Err: "instance not set",
		},
		{
			Title: "realm bad no integ. registers",
			Input: corimCcaRealmBadNoIntegRegs,
			Err: "integrity registers not set",
		},
		{
			Title: "realm bad no raw value",
			Input: corimCcaRealmBadNoRawValue,
			Err: "personalization (raw value) not set",
		},
	}

	common.RunCorimTests(t, tcs)
}
