// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package psa_iot

import (
	"testing"

	"github.com/veraison/services/scheme/common"
)

func TestProfile(t *testing.T) {
	tcs := []common.CorimTestCase{
		{
			Title: "ok",
			Input: corimPsaValid,
		},
		{
			Title: "bad wring class ID type",
			Input: corimPsaBadClass,
			Err: "class ID: expected psa.impl-id, got uuid",
		},
		{
			Title: "bad wring instance type",
			Input: corimPsaBadInstance,
			Err: "instance: expected UEID, got uuid",
		},
		{
			Title: "bad TA no instance",
			Input: corimPsaBadTaNoInstance,
			Err: "instance not set for trust anchor",
		},
		{
			Title: "bad RefVal instance",
			Input: corimPsaBadRefvalInstance,
			Err: "instance set for reference value",
		},
		{
			Title: "bad TA cert",
			Input: corimPsaBadTaCert,
			Err: "trust anchor must be a PKIX base64 key, found: pkix-base64-cert",
		},
		{
			Title: "bad RefVal uint mkey",
			Input: corimPsaBadRefvalMkey,
			Err: "measurement 1 key: expected psa.refval-id, got uint",
		},
		{
			Title: "bad RefVal mval no digests",
			Input: corimPsaBadRefvalMval,
			Err: "measurement 0 value: no digests",
		},
	}

	common.RunCorimTests(t, tcs)
}
