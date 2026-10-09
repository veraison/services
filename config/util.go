// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package config

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var sizeRegexp *regexp.Regexp

// ParseBytesSize parses provided input string into an int64 value. The input
// must be parse-able as decimal integer along with an optional unit specifier.
// The unit specifier can be "KB", "MB", or "GB", where "B" is optional. IEC
// versions (e.g. "KiB") are also supported. The units are always interpreted in
// base 2 (so "KB" is treated as "KiB"). The units are case insensitive. There
// maybe any number of spaces between the value and the unit specifier, as well
// as at the beginning and end of input.
func ParseBytesSize(in string) (int64, error) {
	match := sizeRegexp.FindStringSubmatch(strings.ToLower(in))
	if match == nil {
		return 0, fmt.Errorf("invalid size string: %q", in)
	}

	value, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return 0, err
	}

	multiplier := int64(1)
	switch match[2] {
	case "":
		multiplier = 1
	case "k":
		multiplier = 1 << 10
	case "m":
		multiplier = 1 << 20
	case "g":
		multiplier = 1 << 30
	default:
		// unreachable due to regex matching
		panic(fmt.Sprintf("invalid unit specifier: %s", match[2]))
	}

	ret := value * multiplier

	if value > 1 && multiplier > 1 && ret/value != multiplier {
		return 0, errors.New("overflow")
	}

	return ret, nil
}

func init() {
	sizeRegexp = regexp.MustCompile(`^\s*(\d+)\s*(?:([gmk])i?)?b?\s*$`)
}
