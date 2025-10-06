// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_tpm

import (
	"regexp"
	"strings"

	"github.com/veraison/services/log"
)

// Validation patterns and limits
var (
	// Common potentially dangerous characters that should be escaped or sanitized
	dangerousChars = regexp.MustCompile(`[<>&'"{}()\\]`)
	
	// Pattern to detect suspiciously high frequency of non-word chars
	highFreqSpecials = regexp.MustCompile(`[\W]{20,}`)

	// Allowed character sets for vendor/model strings
	// Includes alphanumeric, common punctuation, and Unicode ranges for various scripts
	allowedChars = regexp.MustCompile(`^[\p{L}\p{N}\p{P}\p{Zs}\t\n]+$`)
)

// ValidationResult represents the outcome of string validation
type ValidationResult struct {
	Valid   bool
	Reasons []string
}

// validateString performs security validation on the input string
func validateString(input string) ValidationResult {
	result := ValidationResult{
		Valid:   true,
		Reasons: []string{},
	}

	// Check for repeated sequences (potential DoS) - simpler approach
	if hasExcessiveRepetition(input) {
		result.Valid = false
		result.Reasons = append(result.Reasons, "excessive repeated sequences detected")
	}

	// Check for high frequency of special characters
	if highFreqSpecials.MatchString(input) {
		result.Valid = false
		result.Reasons = append(result.Reasons, "suspicious frequency of special characters")
	}

	// Validate character set
	if !allowedChars.MatchString(input) {
		result.Valid = false
		result.Reasons = append(result.Reasons, "invalid characters detected")
	}

	// Check character frequency distribution
	if hasAbnormalDistribution(input) {
		result.Valid = false
		result.Reasons = append(result.Reasons, "abnormal character distribution")
	}

	return result
}

// hasExcessiveRepetition checks for excessive repeated characters
func hasExcessiveRepetition(input string) bool {
	if len(input) < 30 {
		return false
	}
	
	maxRepeat := 0
	currentRepeat := 1
	var prevRune rune
	
	for i, r := range input {
		if i > 0 && r == prevRune {
			currentRepeat++
			if currentRepeat > maxRepeat {
				maxRepeat = currentRepeat
			}
		} else {
			currentRepeat = 1
		}
		prevRune = r
	}
	
	// If any character repeats more than 20 times consecutively, flag it
	return maxRepeat > 20
}

// hasAbnormalDistribution checks if the string has an unusually skewed
// distribution of characters (possible encoded/obfuscated content)
func hasAbnormalDistribution(input string) bool {
	if len(input) < 8 {
		return false
	}

	// Count character frequencies
	freqs := make(map[rune]int)
	total := 0
	for _, r := range input {
		freqs[r]++
		total++
	}

	// Check if any character appears with suspicious frequency
	threshold := float64(total) * 0.4 // 40% threshold
	for _, count := range freqs {
		if float64(count) > threshold {
			return true
		}
	}

	return false
}

// sanitizeString performs comprehensive sanitization on input strings to prevent
// potential security issues. It:
// 1. Removes or escapes potentially dangerous characters
// 2. Preserves valid Unicode characters
// 3. Maintains readability of the string
// 4. Returns the sanitized string
func sanitizeString(input string) string {
	// First, validate the input
	validationResult := validateString(input)
	if !validationResult.Valid {
		// If validation fails, apply aggressive sanitization
		input = aggressiveSanitize(input)
	}

	// Replace potentially dangerous characters with their escaped versions
	sanitized := dangerousChars.ReplaceAllStringFunc(input, func(s string) string {
		return "\\" + s
	})

	// Remove any null bytes
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")

	// Remove any control characters except newline and tab
	var builder strings.Builder
	for _, r := range sanitized {
		if !isForbiddenControl(r) {
			builder.WriteRune(r)
		}
	}

	// Log suspicious inputs for security monitoring
	if !validationResult.Valid {
		logSuspiciousInput(input, validationResult.Reasons)
	}

	return builder.String()
}

// isForbiddenControl returns true if the rune is a control character
// that should be removed, excluding commonly used whitespace characters.
func isForbiddenControl(r rune) bool {
	return (r < 32 && r != '\n' && r != '\t') || (r >= 127 && r < 160)
}

// aggressiveSanitize performs more strict sanitization for suspicious inputs
func aggressiveSanitize(input string) string {
	// Remove excessive repetition by limiting consecutive identical characters
	var builder strings.Builder
	var prevRune rune
	repeatCount := 0
	
	for i, r := range input {
		if i > 0 && r == prevRune {
			repeatCount++
			// Allow up to 3 consecutive identical characters
			if repeatCount < 3 {
				builder.WriteRune(r)
			}
		} else {
			repeatCount = 0
			builder.WriteRune(r)
		}
		prevRune = r
	}
	
	input = builder.String()

	// Limit consecutive special characters
	input = highFreqSpecials.ReplaceAllStringFunc(input, func(s string) string {
		if len(s) > 5 {
			return s[:5] // Keep only first 5 special characters
		}
		return s
	})

	// Keep only allowed characters
	builder.Reset()
	for _, r := range input {
		if allowedChars.MatchString(string(r)) {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

// logSuspiciousInput logs potentially malicious input attempts
func logSuspiciousInput(input string, reasons []string) {
	// Use the common logging interface
	log.Warnf("Suspicious TPM vendor/model input detected: %s", strings.Join(reasons, ", "))
	
	// Additional context for security monitoring
	if len(input) > 32 {
		// Log truncated input to avoid log injection
		log.Debugf("Suspicious input (truncated): %q...", input[:32])
	}
}