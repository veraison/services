// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"github.com/veraison/corim/comid"
	"github.com/veraison/services/handler"
)

// IExtractor is the interface that CoRIM plugins need to implement to hook into
// the UnsignedCorimDecoder logics.
// Each extractor consumes a specific CoMID triple and produces a corresponding
// Veraison Endorsement format (or an error).
//
// Note: At the moment the interface is limited by the known use cases.  We
// anticipate that in the future there will to be an extractor for each of the
// defined CoMID triples, plus maybe a way to handle cross-triples checks as
// well as extraction from the "global" CoRIM context.
// See also https://github.com/veraison/services/issues/70
type IExtractor interface {
	RefValExtractor(comid.ValueTriples) ([]*handler.Endorsement, error)
	TaExtractor(comid.KeyTriple) (*handler.Endorsement, error)
	SetProfile(string)
}
