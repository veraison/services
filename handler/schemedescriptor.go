// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"errors"
	"slices"

	"github.com/veraison/services/vts/appraisal"
)

// SchemeVersion represents the versions information for the scheme
type SchemeVersion struct {
	// Major version of the scheme. Changes to the major version indicate
	// compatibility breaks (e.g. dropping of previously-supported input
	// formats, changes to existing entries in the attestation result, etc).
	Major int
	// Minor version of the scheme. Changes to the minor version (without a
	// major version change)indicate backward-compatible changes (e.g.
	// support for new input formats, additions to the attestation result,
	// purely internal changes, etc).
	Minor int
}

// SchemeDescriptor groups together descriptive information about an
// attestation scheme.
type SchemeDescriptor struct {
	// The name of the attestation scheme. This is used to identify the
	// scheme. It also forms a part of the policy ID in the attestation
	// result. This must be unique.
	Name string
	// VersionMajor is the current major version of the scheme (see
	// SchemeVersion above).
	VersionMajor int
	// VersionMinor is the current minor version of the scheme (see
	// SchemeVersion above).
	VersionMinor int
	// CorimProfiles is a list of CoRIM profiles containing endorsements
	// and trust anchors for this scheme. This must not overlap with any
	// other registered scheme.
	CorimProfiles []string
	// EvidenceMediaTypes is the list of attesation evidence media types
	// handled by this scheme. This must not overlap with any other
	// registered scheme.
	EvidenceMediaTypes []string
}

func (o *SchemeDescriptor) Validate() error {
	if o.Name == "" {
		return errors.New("name not set")
	}

	if len(o.CorimProfiles) == 0 {
		return errors.New("CoRIM profiles not set")
	}

	if len(o.EvidenceMediaTypes)  == 0 {
		return errors.New("evidence media types not set")
	}

	return nil
}

func (o *SchemeDescriptor) EvidenceIsSupported(ev *appraisal.Evidence) bool {
	return slices.Contains(o.EvidenceMediaTypes, ev.MediaType)
}

func (o *SchemeDescriptor) Version() SchemeVersion {
	return SchemeVersion{Major: o.VersionMajor, Minor: o.VersionMinor}
}

