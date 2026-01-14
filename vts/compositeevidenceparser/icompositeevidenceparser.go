// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package compositeevidenceparser

type ICompositeEvidenceParser interface {
	// Parse returns a list of component evidence payloads together with
	// relevant identifying metadata.
	Parse(evidence []byte) ([]ComponentEvidence, error)
}

type ComponentEvidence struct {
	label     string
	data      []byte
	mediaType string
}
