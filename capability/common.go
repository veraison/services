// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package capability

import (
	"time"

	"go.uber.org/zap"
)

// ParseCacheMaxAge parses a "max age" configuration string expressed in time
// units (e.g., "1h", "30m") and returns a time.Duration.
// If the input string is empty, negative or invalid, the function returns the
// provided default duration and logs a warning message.
// A duration string is a possibly signed sequence of decimal numbers, each with
// optional fraction and a unit suffix, such as "300s", "1.5h" or "2h45m".
// Valid time units are "h", "m", "s".
func ParseCacheMaxAge(s string, dflt time.Duration, logger *zap.SugaredLogger) time.Duration {
	if s == "" {
		return dflt
	}

	ma, err := time.ParseDuration(s)
	if err != nil {
		logger.Warnf("invalid .well-known cache max age: %v. resetting to default (%v)", err, dflt)
		return dflt
	}

	if ma < 0 {
		logger.Warnf("negative .well-known cache max age: %v. resetting to default (%v)", err, dflt)
		return dflt
	}

	return ma
}
