// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_tpm

import (
	"fmt"
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/corim/comid"
)

func TestTaAttrs_RealWorldVendors(t *testing.T) {
	// Test cases based on real-world TPM vendors and models
	testCases := []struct {
		name      string
		vendor    string
		model     string
		wantError bool
	}{
		{
			name:   "Infineon",
			vendor: "Infineon Technologies AG",
			model:  "SLB 9670 TPM2.0",
		},
		{
			name:   "STMicroelectronics",
			vendor: "STMicroelectronics",
			model:  "ST33TPHF2ESPI TPM 2.0",
		},
		{
			name:   "Nuvoton",
			vendor: "Nuvoton Technology Corp.",
			model:  "NPCT750 TPM2.0",
		},
		{
			name:   "Nationz",
			vendor: "Nationz Technologies Inc.",
			model:  "TPM 2.0 Z32H330",
		},
		{
			name:   "Intel",
			vendor: "Intel Corporation",
			model:  "Intel® PTT",
		},
		{
			name:   "AMD",
			vendor: "Advanced Micro Devices, Inc.",
			model:  "AMD fTPM",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id := ID{
				class:    "test-class",
				instance: "test-instance",
			}

			key := &comid.CryptoKey{
				Parameters: map[string]interface{}{
					"vendor": tc.vendor,
					"model":  tc.model,
				},
				Value: &comid.TaggedPKIXBase64Key{PK: "test-key-data"},
			}

			attrs, err := makeTaAttrs(id, key)
			if tc.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			var taAttr TaAttr
			err = json.Unmarshal(attrs, &taAttr)
			require.NoError(t, err)

			assert.Equal(t, tc.vendor, *taAttr.Vendor)
			assert.Equal(t, tc.model, *taAttr.Model)
		})
	}
}

func TestTaAttrs_InternationalVendors(t *testing.T) {
	// Test cases with international vendor names and models
	testCases := []struct {
		name      string
		vendor    string
		model     string
		wantError bool
	}{
		{
			name:   "Japanese Vendor",
			vendor: "富士通株式会社",
			model:  "FUJITSU TPM v2.0",
		},
		{
			name:   "Chinese Vendor",
			vendor: "联想集团有限公司",
			model:  "ThinkPad TPM 2.0",
		},
		{
			name:   "Korean Vendor",
			vendor: "삼성전자",
			model:  "Samsung TPM2.0",
		},
		{
			name:   "Russian Vendor",
			vendor: "Компания",
			model:  "ТПМ-2000",
		},
		{
			name:   "Mixed Scripts",
			vendor: "Fujitsu富士通",
			model:  "TPM v2.0 型号",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id := ID{
				class:    "test-class",
				instance: "test-instance",
			}

			key := &comid.CryptoKey{
				Parameters: map[string]interface{}{
					"vendor": tc.vendor,
					"model":  tc.model,
				},
				Value: &comid.TaggedPKIXBase64Key{PK: "test-key-data"},
			}

			attrs, err := makeTaAttrs(id, key)
			require.NoError(t, err)

			var taAttr TaAttr
			err = json.Unmarshal(attrs, &taAttr)
			require.NoError(t, err)

			assert.Equal(t, tc.vendor, *taAttr.Vendor)
			assert.Equal(t, tc.model, *taAttr.Model)
		})
	}
}

func TestTaAttrs_EdgeCases(t *testing.T) {
	testCases := []struct {
		name      string
		vendor    string
		model     string
		wantError bool
	}{
		{
			name:   "Maximum Length",
			vendor: strings.Repeat("V", 1024),
			model:  strings.Repeat("M", 1024),
		},
		{
			name:      "Exceeds Maximum Length",
			vendor:    strings.Repeat("V", 1025),
			model:     strings.Repeat("M", 1025),
			wantError: true,
		},
		{
			name:   "Single Character",
			vendor: "V",
			model:  "M",
		},
		{
			name:   "Space Only",
			vendor: "   ",
			model:  "   ",
		},
		{
			name:   "Mixed Spaces",
			vendor: "   Vendor   Name   ",
			model:  "   Model   Type   ",
		},
		{
			name:   "Special Characters",
			vendor: "Vendor-Name & Co., Ltd.",
			model:  "TPM-2000 (Gen2) v2.0",
		},
		{
			name:   "With Trademark Symbols",
			vendor: "Company™",
			model:  "TPM® 2.0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id := ID{
				class:    "test-class",
				instance: "test-instance",
			}

			key := &comid.CryptoKey{
				Parameters: map[string]interface{}{
					"vendor": tc.vendor,
					"model":  tc.model,
				},
				Value: &comid.TaggedPKIXBase64Key{PK: "test-key-data"},
			}

			attrs, err := makeTaAttrs(id, key)
			if tc.wantError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			var taAttr TaAttr
			err = json.Unmarshal(attrs, &taAttr)
			require.NoError(t, err)

			// For non-error cases, verify the strings are properly trimmed and sanitized
			if !tc.wantError {
				assert.Equal(t, strings.TrimSpace(tc.vendor), *taAttr.Vendor)
				assert.Equal(t, strings.TrimSpace(tc.model), *taAttr.Model)
			}
		})
	}
}

func TestTaAttrs_SecurityCases(t *testing.T) {
	testCases := []struct {
		name           string
		vendor         string
		model          string
		wantError      bool
		checkSanitized bool
	}{
		{
			name:           "HTML Injection",
			vendor:         "<script>alert('xss')</script>",
			model:         "<img src=x onerror=alert('xss')>",
			checkSanitized: true,
		},
		{
			name:           "SQL Injection",
			vendor:         "1' OR '1'='1",
			model:         "1'; DROP TABLE users--",
			checkSanitized: true,
		},
		{
			name:           "Command Injection",
			vendor:         "$(rm -rf /)",
			model:         "`rm -rf /`",
			checkSanitized: true,
		},
		{
			name:           "Path Traversal",
			vendor:         "../../../etc/passwd",
			model:         "..\\..\\..\\windows\\system32",
			checkSanitized: true,
		},
		{
			name:      "Null Bytes",
			vendor:    "Vendor\x00Name",
			model:     "Model\x00Type",
			wantError: true,
		},
		{
			name:      "Control Characters",
			vendor:    "Bad\x01Vendor\x02Name",
			model:     "Bad\x01Model\x02Type",
			wantError: true,
		},
		{
			name:           "Mixed Valid and Invalid",
			vendor:         "Valid Name <script>alert(1)</script>",
			model:         "Valid Model `rm -rf /`",
			checkSanitized: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id := ID{
				class:    "test-class",
				instance: "test-instance",
			}

			key := &comid.CryptoKey{
				Parameters: map[string]interface{}{
					"vendor": tc.vendor,
					"model":  tc.model,
				},
				Value: &comid.TaggedPKIXBase64Key{PK: "test-key-data"},
			}

			attrs, err := makeTaAttrs(id, key)
			if tc.wantError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			var taAttr TaAttr
			err = json.Unmarshal(attrs, &taAttr)
			require.NoError(t, err)

			if tc.checkSanitized {
				// Check that dangerous characters are escaped or removed
				assert.NotEqual(t, tc.vendor, *taAttr.Vendor, "Dangerous vendor string should be sanitized")
				assert.NotEqual(t, tc.model, *taAttr.Model, "Dangerous model string should be sanitized")
				
				// Verify the sanitized strings don't contain dangerous patterns
				dangerousPatterns := []string{"<script>", "alert", "rm -rf", "../", "\\.\\.\\"}
				for _, pattern := range dangerousPatterns {
					assert.NotContains(t, *taAttr.Vendor, pattern, "Sanitized vendor should not contain %s", pattern)
					assert.NotContains(t, *taAttr.Model, pattern, "Sanitized model should not contain %s", pattern)
				}
			}
		})
	}
}

func TestTaAttrs_Combinations(t *testing.T) {
	// Test various combinations of vendor/model presence and content
	testCases := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "Vendor Only",
			params: map[string]interface{}{
				"vendor": "Test Vendor",
			},
		},
		{
			name: "Model Only",
			params: map[string]interface{}{
				"model": "Test Model",
			},
		},
		{
			name:   "Neither Present",
			params: map[string]interface{}{},
		},
		{
			name: "Multiple Parameters",
			params: map[string]interface{}{
				"vendor":     "Test Vendor",
				"model":      "Test Model",
				"extra_key": "should be ignored",
				"version":   "1.0",
			},
		},
		{
			name: "With Version Info",
			params: map[string]interface{}{
				"vendor": "Test Vendor",
				"model":  "TPM 2.0",
				"fw_version": "1.2.3.4",
				"hw_version": "5.6.7.8",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id := ID{
				class:    "test-class",
				instance: "test-instance",
			}

			key := &comid.CryptoKey{
				Parameters: tc.params,
				Value:     &comid.TaggedPKIXBase64Key{PK: "test-key-data"},
			}

			attrs, err := makeTaAttrs(id, key)
			if tc.wantError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			var taAttr TaAttr
			err = json.Unmarshal(attrs, &taAttr)
			require.NoError(t, err)

			// Verify vendor presence matches input
			if v, ok := tc.params["vendor"]; ok {
				require.NotNil(t, taAttr.Vendor)
				assert.Equal(t, v.(string), *taAttr.Vendor)
			} else {
				assert.Nil(t, taAttr.Vendor)
			}

			// Verify model presence matches input
			if m, ok := tc.params["model"]; ok {
				require.NotNil(t, taAttr.Model)
				assert.Equal(t, m.(string), *taAttr.Model)
			} else {
				assert.Nil(t, taAttr.Model)
			}
		})
	}
}

func TestSanitize_CharacterTypes(t *testing.T) {
	// Test handling of different character types and categories
	testCases := []struct {
		name     string
		input    string
		validate func(t *testing.T, input, sanitized string)
	}{
		{
			name:  "Unicode Categories",
			input: "Lu:A Ll:a Lt:ǅ Lm:ʰ Lo:豈 Nl:Ⅰ Nd:1 Pc:_ Pd:- Ps:( Pe:) Pi:« Pf:» Po:! Sm:+ Sc:£ Sk:^ So:© Zs:  Zl:\u2028 Zp:\u2029",
			validate: func(t *testing.T, input, sanitized string) {
				// Verify that valid Unicode categories are preserved
				for _, r := range sanitized {
					assert.True(t, unicode.IsLetter(r) || 
					          unicode.IsNumber(r) || 
					          unicode.IsPunct(r) || 
					          unicode.IsSymbol(r) ||
					          unicode.IsSpace(r),
					          "Character %q should be preserved", r)
				}
			},
		},
		{
			name:  "Diacritical Marks",
			input: "áéíóúÁÉÍÓÚñÑüÜ",
			validate: func(t *testing.T, input, sanitized string) {
				assert.Equal(t, input, sanitized, "Diacritical marks should be preserved")
			},
		},
		{
			name:  "Technical Symbols",
			input: "±∑∆∏∐∂∅∈∉√∛∜∝∞∟∠∡∢∣",
			validate: func(t *testing.T, input, sanitized string) {
				assert.Equal(t, input, sanitized, "Technical symbols should be preserved")
			},
		},
		{
			name:  "Currency Symbols",
			input: "₠₡₢₣₤₥₦₧₨₩₪₫€₭₮₯",
			validate: func(t *testing.T, input, sanitized string) {
				assert.Equal(t, input, sanitized, "Currency symbols should be preserved")
			},
		},
		{
			name:  "Mixed Scripts",
			input: "Latin:ABC Cyrillic:АБВ Greek:ΑΒΓ Hebrew:אבג Arabic:ابت",
			validate: func(t *testing.T, input, sanitized string) {
				assert.Equal(t, input, sanitized, "Different scripts should be preserved")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sanitized := sanitizeString(tc.input)
			tc.validate(t, tc.input, sanitized)
		})
	}
}