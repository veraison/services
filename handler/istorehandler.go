// Copyright 2024-2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"github.com/veraison/services/plugin"
	"github.com/veraison/services/proto"
)

// IStoreHandler defines the interfaces for creating and obtaining keys
// to access objects in the Veraison storage layer.
// This includes obtaining Trust Anchor IDs from evidence and synthesizing
// Reference Value and TrustAnchor keys from endorsements
type IStoreHandler interface {
	plugin.IPluggable

	// GetTrustAnchorIDs returns a slice of trust anchor identifiers used to
	// retrieve the trust anchors associated with this token. The trust anchors
	// may be necessary to validate the entire token and/or extract its claims
	// (if it is encrypted).
	GetTrustAnchorIDs(token *proto.AttestationToken) ([]string, error)

	// GetRefValueIDs returns a slice of identifiers used to retrieve
	// reference values for an attestation scheme, using the claims
	// extracted from attestation token and the associated trust anchors.
	GetRefValueIDs(
		tenantID string,
		trustAnchors []string,
		claims map[string]interface{},
	) ([]string, error)

	// SynthKeysFromRefValue synthesizes lookup key(s) for the
	// provided reference value endorsement.
	SynthKeysFromRefValue(tenantID string, refVal *Endorsement) ([]string, error)

	// SynthKeysFromTrustAnchor synthesizes lookup key(s) for the provided
	// trust anchor.
	SynthKeysFromTrustAnchor(tenantID string, ta *Endorsement) ([]string, error)

	// SynthCoservQueryKeys synthesizes lookup keys for the supplied CoSERV
	// environment selector.
	SynthCoservQueryKeys(tenantID string, query string) ([]string, error)
}
