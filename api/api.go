// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package api

import "mime"

// NormalizeMediaType validates the supplied media type (including any
// parameters) and returns it in normalized form, i.e., with lowercase type,
// subtype and, optionally, parameter name.  An error is returned if the
// supplied media type is invalid.
func NormalizeMediaType(mt string) (string, error) {
	m, p, err := mime.ParseMediaType(mt)
	if err != nil {
		return "", err
	}

	return mime.FormatMediaType(m, p), nil
}
