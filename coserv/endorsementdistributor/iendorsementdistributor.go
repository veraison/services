// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package endorsementdistributor

import "github.com/veraison/services/proto"

// IEndorsementDistributor provides the bridge between the coserv REST API
// handler and VTS gRPC service endpoints relevant to endorsement distribution.
type IEndorsementDistributor interface {
	// SupportedMediaTypes returns a list of supported media types
	SupportedMediaTypes() ([]string, error)

	// GetEndorsements retrieves endorsements matching the given query
	// for the given tenantID, formatted as specified by mediaType
	// (which must be one of the media types returned by SupportedMediaTypes())
	//
	// If no endorsements can be found matching the query, an empty result set
	// is returned with no error.
	GetEndorsements(tenantID string, query string, mediaType string) ([]byte, error)

	// GetPublicKey returns the public key used to sign CoSERV responses
	// (if any). If no signing is performed, nil is returned.
	GetPublicKey() (*proto.PublicKey, error)
}
