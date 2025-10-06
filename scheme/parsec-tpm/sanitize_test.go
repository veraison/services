// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_tpm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateString(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		wantValid   bool
		wantReasons []string
	}{
		{
			name:      "valid vendor name",
			input:     "Acme Corporation, Ltd.",
			wantValid: true,
		},
		{
			name:        "repeated sequence attack",
			input:       "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
			wantValid:   false,
			wantReasons: []string{"excessive repeated sequences detected", "abnormal character distribution"},
		},
		{
			name:        "special character attack",
			input:       "!@#$%^&*()!@#$%^&*()!@#$%^&*()",
			wantValid:   false,
			wantReasons: []string{"suspicious frequency of special characters", "invalid characters detected"},
		},
		{
			name:        "invalid character attack",
			input:       "Company\x00Name\x01Corp",
			wantValid:   false,
			wantReasons: []string{"invalid characters detected"},
		},
		{
			name:        "skewed distribution attack",
			input:       strings.Repeat("X", 100),
			wantValid:   false,
			wantReasons: []string{"excessive repeated sequences detected", "abnormal character distribution"},
		},
		{
			name:      "valid international name",
			input:     "富士通株式会社",
			wantValid: true,
		},
		{
			name:      "valid mixed content",
			input:     "Hewlett-Packard (HP) 株式会社",
			wantValid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validateString(tc.input)
			assert.Equal(t, tc.wantValid, result.Valid)
			if !tc.wantValid {
				assert.Equal(t, tc.wantReasons, result.Reasons)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal text",
			input:    "Normal vendor name",
			expected: "Normal vendor name",
		},
		{
			name:     "html injection attempt",
			input:    "Vendor<script>alert('xss')</script>",
			expected: "Vendorscriptalert\\(\\'xss\\'\\)/script",
		},
		{
			name:     "sql injection attempt",
			input:    "Vendor'; DROP TABLE users;--",
			expected: "Vendor\\'; DROP TABLE users;--",
		},
		{
			name:     "null byte injection",
			input:    "Vendor\x00Name",
			expected: "VendorName",
		},
		{
			name:     "control characters",
			input:    "Vendor\x01Name\x02",
			expected: "VendorName",
		},
		{
			name:     "allowed whitespace",
			input:    "Vendor\nName\tCorp",
			expected: "Vendor\nName\tCorp",
		},
		{
			name:     "unicode characters",
			input:    "製造元株式会社",
			expected: "製造元株式会社",
		},
		{
			name:     "mixed content",
			input:    "Vendor & Co<> \x00株式会社\n",
			expected: "Vendor \\& Co 株式会社\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizeString(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}