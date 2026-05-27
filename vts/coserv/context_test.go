// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package coserv

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_parseDuration(t *testing.T) {
	testCases := []struct {
		title    string
		text     string
		expected time.Duration
		err      string
	}{
		{
			title:    "ok spaces",
			text:     " 2 min 7 secs\n",
			expected: 127 * time.Second,
		},
		{
			title:    "ok no spaces",
			text:     "1d2h3m4s",
			expected: 26*time.Hour + 184*time.Second,
		},
		{
			title:    "ok repeated",
			text:     " 2 min 2 min\n",
			expected: 4 * time.Minute,
		},
		{
			title: "bad no matches",
			text:  " foo",
			err:   "could not match time specifiers",
		},
		{
			title:    "bad unrecognized middle",
			text:     " 2 min foo 1 sec",
			expected: 4 * time.Minute,
			err:      "unrecognized text between indexes 7 and 10",
		},
		{
			title:    "bad unrecognized trailing",
			text:     " 2 min 1 secfoo",
			expected: 4 * time.Minute,
			err:      "unrecognized trailing text",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			duration, err := parseDuration(tc.text)
			if tc.err == "" {
				assert.NoError(t, err)
				assert.EqualValues(t, tc.expected, duration)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}
