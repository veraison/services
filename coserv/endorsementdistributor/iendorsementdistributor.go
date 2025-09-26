// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package endorsementdistributor

type IEndorsementDistributor interface {
	// TODO
	// GetVTSState() (*proto.ServiceState, error)
	// IsSupportedProfile(profile string) (bool, error)

	// SupportedProfiles returns a list of supported profile names
	SupportedProfiles() ([]string, error)

	// GetEndorsements retrieves endorsements matching the given query
	// for the given tenantID, formatted as specified by resultMediaType
	// (which must be one of the media types returned by SupportedMediaTypes())
	//
	// If no endorsement can be found matching the query, an empty result set
	// is returned with no error.
	GetEndorsements(tenantID string, query string, resultMediaType string) ([]byte, error)
}
