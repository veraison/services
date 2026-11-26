// Copyright 2024-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package util

import "mime"

// NormalizeMediaType validates the supplied media type (including any
// parameters) and returns it in normalized form, i.e., with lowercase type,
// subtype and, optionally, parameter name. If the media type is invalid,
// it is returned unchangned along with an error
func NormalizeMediaType(mt string) (string, error) {
	m, p, err := mime.ParseMediaType(mt)
	if err != nil {
		return mt, err
	}

	return mime.FormatMediaType(m, p), nil
}
