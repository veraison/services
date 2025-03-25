// Copyright 2025 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package endorsementdistributor

type IEndorsementDistributor interface {
	// TODO
	// GetVTSState() (*proto.ServiceState, error)
	// IsSupportedProfile(profile string) (bool, error)
	// SupportedProfiles() ([]string, error)

	// GetEndorsements
	GetEndorsements(tenantID string, query string, resultMediaType string) ([]byte, error)
}
