// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package compositeevidenceparser

import (
	"fmt"
	"mime"
)

// present MediaType to Composite Parser is maintained locally within the compositeevidenceparser package
var mtToCeParser = map[string]ICompositeEvidenceParser{
	"application/cmw-collection+cbor": &cmwParser{},
	"application/cmw-collection+json": &cmwParser{},
}

func GetParserFromMediaType(mt string) (ICompositeEvidenceParser, error) {
	// Check if its a valid mediaType
	if _, _, err := mime.ParseMediaType(mt); err != nil {
		return nil, fmt.Errorf("bad media type: %w", err)
	}
	switch mt {
	case "application/cmw-collection+cbor", "application/cmw-collection+json":
		return mtToCeParser[mt], nil
	default:
		return nil, fmt.Errorf("unsupported media type:%s", mt)
	}
}
