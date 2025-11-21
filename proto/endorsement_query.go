// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package proto

// GetEndorsementsRequest is the request message for GetEndorsements RPC
type GetEndorsementsRequest struct {
	// Optional key prefix to filter endorsements. If empty, returns all endorsements.
	KeyPrefix string `json:"key_prefix,omitempty"`
	// Type of endorsement to retrieve: "trust-anchor", "reference-value", or "all"
	EndorsementType string `json:"endorsement_type,omitempty"`
}

// EndorsementEntry represents a single endorsement entry
type EndorsementEntry struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

// GetEndorsementsResponse is the response message for GetEndorsements RPC
type GetEndorsementsResponse struct {
	Endorsements []*EndorsementEntry `json:"endorsements"`
	Status       *Status             `json:"status"`
}

// DeleteEndorsementsRequest is the request message for DeleteEndorsements RPC
type DeleteEndorsementsRequest struct {
	// Key or key prefix to delete
	Key string `json:"key"`
	// Type of endorsement to delete: "trust-anchor", "reference-value", or "all"
	EndorsementType string `json:"endorsement_type,omitempty"`
}

// DeleteEndorsementsResponse is the response message for DeleteEndorsements RPC
type DeleteEndorsementsResponse struct {
	DeletedCount int32   `json:"deleted_count"`
	Status       *Status `json:"status"`
}
