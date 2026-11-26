// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package endorsement

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestStoreConfigFromViper(t *testing.T) {
	testCases := []struct {
		title string
		input *viper.Viper
		err   string
	}{
		{
			title: "err empty",
			input: viper.New(),
			err:   "'' has unset fields: active-plugins, coserv",
		},
		{
			title: "err empty coserv",
			input: func() *viper.Viper {
				v := viper.New()
				v.Set("active-plugins", []string{"corim-store"})
				v.Set("coserv", viper.New())
				return v
			}(),
			err: "'coserv' has unset fields: max-expiry, signer",
		},
		{
			title: "err bad signer key",
			input: func() *viper.Viper {
				v := viper.New()
				v.Set("active-plugins", []string{"corim-store"})
				v.Set("coserv.signer.alg", "ES256")
				v.Set("coserv.signer.key", "./testdata/nonexistent.jwk")
				v.Set("coserv.max-expiry", "5m")
				return v
			}(),
			err: "signer: loading CoSERV signer key",
		},
		{
			title: "err bad signer alg",
			input: func() *viper.Viper {
				v := viper.New()
				v.Set("active-plugins", []string{"corim-store"})
				v.Set("coserv.signer.alg", "BAD-ALG")
				v.Set("coserv.signer.key", "./testdata/skey.jwk")
				v.Set("coserv.max-expiry", "5m")
				return v
			}(),
			err: `"BAD-ALG" is not a valid value for "coserv-signer.alg"`,
		},
		{
			title: "err bad max-expiry",
			input: func() *viper.Viper {
				v := viper.New()
				v.Set("active-plugins", []string{"corim-store"})
				v.Set("coserv.signer.alg", "ES256")
				v.Set("coserv.signer.key", "./testdata/skey.jwk")
				v.Set("coserv.max-expiry", "not-a-duration")
				return v
			}(),
			err: `max expiry: could not match time specifiers in "not-a-duration"`,
		},
		{
			title: "err bad plugin params",
			input: func() *viper.Viper {
				v := viper.New()
				v.Set("active-plugins", []string{"corim-store"})
				v.Set("coserv.signer.alg", "ES256")
				v.Set("coserv.signer.key", "./testdata/skey.jwk")
				v.Set("coserv.max-expiry", "5m")
				v.Set("plugin-parameters.corim-store.bad-param", struct{}{})
				return v
			}(),
			err: `plugin "corim-store" params: invalid: value for "bad-param"`,
		},
		{
			title: "err no active plugins",
			input: func() *viper.Viper {
				v := viper.New()
				v.Set("active-plugins", []string{})
				v.Set("coserv.signer.alg", "ES256")
				v.Set("coserv.signer.key", "./testdata/skey.jwk")
				v.Set("coserv.max-expiry", "5m")
				v.Set("plugin-parameters.corim-store.dbms", "sqlite3")
				v.Set("plugin-parameters.corim-store.dsn", "file::memory:?cache=shared")
				v.Set("plugin-parameters.corim-store.trace-sql", false)
				return v
			}(),
			err: "endorsement-store config: no active plugins specified",
		},
		{
			title: "ok",
			input: func() *viper.Viper {
				v := viper.New()
				v.Set("active-plugins", []string{"corim-store"})
				v.Set("coserv.signer.alg", "ES256")
				v.Set("coserv.signer.key", "./testdata/skey.jwk")
				v.Set("coserv.max-expiry", "5m")
				v.Set("plugin-parameters.corim-store.dbms", "sqlite3")
				v.Set("plugin-parameters.corim-store.dsn", "file::memory:?cache=shared")
				v.Set("plugin-parameters.corim-store.trace-sql", false)
				return v
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			_, err := StoreConfigFromViper(tc.input)

			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}

		})
	}
}

func TestParseDuration(t *testing.T) {
	minute := time.Minute
	testCases := []struct {
		title    string
		input    string
		expected time.Duration
		err      string
	}{
		{
			title:    "ok s",
			input:    "1s",
			expected: time.Second,
		},
		{
			title:    "ok sec",
			input:    "1sec",
			expected: time.Second,
		},
		{
			title:    "ok second",
			input:    "1second",
			expected: time.Second,
		},
		{
			title:    "ok seconds",
			input:    "2seconds",
			expected: 2 * time.Second,
		},
		{
			title:    "ok m",
			input:    "1m",
			expected: minute,
		},
		{
			title:    "ok min",
			input:    "1min",
			expected: minute,
		},
		{
			title:    "ok minute",
			input:    "1minute",
			expected: minute,
		},
		{
			title:    "ok minutes",
			input:    "2minutes",
			expected: 2 * minute,
		},
		{
			title:    "ok h",
			input:    "1h",
			expected: time.Hour,
		},
		{
			title:    "ok hour",
			input:    "1hour",
			expected: time.Hour,
		},
		{
			title:    "ok hours",
			input:    "2hours",
			expected: 2 * time.Hour,
		},
		{
			title:    "ok d",
			input:    "1d",
			expected: 24 * time.Hour,
		},
		{
			title:    "ok day",
			input:    "1day",
			expected: 24 * time.Hour,
		},
		{
			title:    "ok days",
			input:    "2days",
			expected: 48 * time.Hour,
		},
		{
			title:    "ok uppercase",
			input:    "1H",
			expected: time.Hour,
		},
		{
			title:    "ok whitespace",
			input:    "  1 h  ",
			expected: time.Hour,
		},
		{
			title:    "ok compound",
			input:    "1h30m",
			expected: 90 * minute,
		},
		{
			title:    "ok compound with spaces",
			input:    "1 hour 30 minutes",
			expected: 90 * minute,
		},
		{
			title: "err no match",
			input: "not-a-duration",
			err:   `could not match time specifiers in "not-a-duration"`,
		},
		{
			title: "err unrecognized text between specifiers",
			input: "1h foo 30m",
			err:   "unrecognized text between indexes",
		},
		{
			title: "err unrecognized trailing text",
			input: "1h foo",
			err:   "unrecognized trailing text",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			actual, err := parseDuration(tc.input)

			if tc.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			} else {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}
