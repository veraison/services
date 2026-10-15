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
			Err:   "environment.class is required",
		},
		{
			Title: "platform bad TA no instance",
			Input: corimCcaPlatformBadTaNoInstance,
			Err:   "environment.instance (instance-id) is required",
		},
		{
			Title: "platform bad TA bytes instance",
			Input: corimCcaPlatformBadTaInstance,
			Err:   "instance-id must be of type 'ueid', got 'bytes'",
		},
		{
			Title: "platform bad TA cert",
			Input: corimCcaPlatformBadTaCert,
			Err:   "verification-key must be of type 'pkix-base64-key', got 'pkix-base64-cert'",
		},
		{
			Title: "platform bad RefVal no mkey",
			Input: corimCcaPlatformBadRefvalNoMkey,
			Err:   "mkey is mandatory but not set",
		},
		{
			Title: "platform bad RefVal uint mkey",
			Input: corimCcaPlatformBadRefvalMkey,
			Err:   "mkey must be of type 'string', got 'uint'",
		},
		{
			Title: "platform bad RefVal invalid string mkey",
			Input: corimCcaPlatformBadRefvalMkeyString,
			Err:   "invalid mkey \"cca.bad-component\"",
		},
		{
			Title: "platform bad RefVal malformed cryptokeys",
			Input: corimCcaPlatformBadRefvalCryptokeys,
			Err:   "cryptokeys (signer-id) must be of type 'bytes'",
		},
		{
			Title: "platform bad RefVal no digest",
			Input: corimCcaPlatformBadRefvalNoDigests,
			Err:   "digests field is mandatory but not set",
		},
		{
			Title: "platform bad RefVal no raw value",
			Input: corimCcaPlatformBadRefvalNoRawValue,
			Err:   "cca.platform-config raw value error: raw-value is mandatory",
		},
		{
			Title: "realm ok",
			Input: corimCcaRealmValid,
		},
		{
			Title: "realm bad instance",
			Input: corimCcaRealmBadInstance,
			Err:   "RIM must be of type 'bytes', got 'uuid'",
		},
		{
			Title: "realm bad no instance",
			Input: corimCcaRealmBadNoInstance,
			Err:   "environment.class is required for CCA Realm profile",
		},
		{
			Title: "realm bad no rim",
			Input: corimCcaRealmBadNoRim,
			Err:   "RIM (cca.rim) measurement is mandatory but not found",
		},
		{
			Title: "realm bad no raw value",
			Input: corimCcaRealmBadNoRawValue,
			Err:   "raw-value is mandatory for cca.rpv",
		},
		{
			Title: "realm bad no integ regs",
			Input: corimCcaRealmBadNoIntegRegs,
			Err:   "digests field is mandatory but not set",
		},
	}

	common.RunCorimTests(t, tcs)
}
