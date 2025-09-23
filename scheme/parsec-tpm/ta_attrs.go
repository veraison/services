// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package parsec_tpm

import (
	"encoding/json"
	"fmt"

	"github.com/veraison/corim/comid"
)

func makeTaAttrs(id ID, key *comid.CryptoKey) (json.RawMessage, error) {
	taAttr := TaAttr{
		ClassID: &id.class,
		InstID: &id.instance,
	}

	// Extract and validate optional vendor and model from metadata.
	// Security and validation rules:
	// 1. Only string values are accepted
	// 2. Empty strings are allowed but trimmed
	// 3. Maximum length of 1024 characters for security
	// 4. Unicode and special characters are supported
	// 5. Strings are sanitized to prevent injection
	// 6. If validation fails, the field is omitted (no error)
	if meta, ok := key.Parameters["vendor"].(string); ok {
		// Trim and validate length
		meta = strings.TrimSpace(meta)
		if len(meta) <= 1024 {
			// Basic sanitization
			meta = sanitizeString(meta)
			taAttr.Vendor = &meta
		}
	}
	if meta, ok := key.Parameters["model"].(string); ok {
		// Trim and validate length
		meta = strings.TrimSpace(meta)
		if len(meta) <= 1024 {
			// Basic sanitization
			meta = sanitizeString(meta)
			taAttr.Model = &meta
		}
	}

	// Convert the key to a base64 encoded string
	pubkey, err := comidKeyToPEM(key)
	if err != nil {
		return nil, fmt.Errorf("converting key to PEM: %w", err)
	}
	taAttr.VerifKey = &pubkey

	attrs, err := json.Marshal(taAttr)
	if err != nil {
		return nil, fmt.Errorf("marshaling attributes: %w", err)
	}

	return attrs, nil
}