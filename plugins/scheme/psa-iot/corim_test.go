// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package psa_iot

import (
	"testing"

	"github.com/veraison/services/pkg/util"
)

func TestProfile(t *testing.T) {
	tcs := []util.CorimTestCase{
		{
			Title: "ok",
			Input: corimPsaValid,
		},
		{
			Title: "bad wring class ID type",
			Input: corimPsaBadClass,
			Err:   "implementation-id must be of type 'bytes', got 'uuid'",
		},
		{
			Title: "bad wring instance type",
			Input: corimPsaBadInstance,
			Err:   "instance-id must be of type 'ueid', got 'uuid'",
		},
		{
			Title: "bad TA no instance",
			Input: corimPsaBadTaNoInstance,
			Err:   "environment.instance (instance-id) is required",
		},
		{
			Title: "bad TA cert",
			Input: corimPsaBadTaCert,
			Err:   "verification-key must be of type 'pkix-base64-key', got 'pkix-base64-cert'",
		},
		{
			Title: "bad RefVal uint mkey",
			Input: corimPsaBadRefvalMkey,
			Err:   "mkey must be of type 'string', got 'uint'",
		},
	}

	util.RunCorimTests(t, tcs)
}
