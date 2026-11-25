// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBytesSize(t *testing.T) {
	testCases := []struct {
		title    string
		in       string
		expected int64
		err      string
	}{
		{
			title:    "ok no unit",
			in:       "1234",
			expected: 1234,
		},
		{
			title:    "ok B",
			in:       "1234B",
			expected: 1234,
		},
		{
			title:    "ok k",
			in:       "1234k",
			expected: 1234 * 1024,
		},
		{
			title:    "ok MB",
			in:       "1234MB",
			expected: 1234 * 1024 * 1024,
		},
		{
			title:    "ok MiB",
			in:       "1234MiB",
			expected: 1234 * 1024 * 1024,
		},
		{
			title:    "ok KB with spaces",
			in:       "  1234 KB  ",
			expected: 1234 * 1024,
		},
		{
			title:    "ok gb lower case",
			in:       "1234 gb",
			expected: 1234 * 1024 * 1024 * 1024,
		},
		{
			title: "err bad format",
			in:    "1234 KiB 7",
			err:   `invalid size string: "1234 KiB 7"`,
		},
		{
			title: "err bad uint",
			in:    "1234 ZiB",
			err:   `invalid size string: "1234 ZiB"`,
		},
		{
			title: "err empty",
			in:    "",
			err:   `invalid size string: ""`,
		},
		{
			title: "err too big",
			in:    "99999999999999999999",
			err:   "value out of range",
		},
		{
			title: "err overflow",
			in:    "99999999999999999MB",
			err:   "overflow",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			ret, err := ParseBytesSize(tc.in)
			if tc.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, ret)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}
