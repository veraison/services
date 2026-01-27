// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package compositeevidenceparser

type ICompositeEvidenceParser interface {
	// SupportedMediaTypes returns a list of supported Collection Media Types
	// supported by the Parser
	SupportedMediaTypes() []string
	// Parse returns a list of component evidence payloads together with
	// relevant identifying metadata.
	Parse(evidence []byte) ([]ComponentEvidence, error)
}

type ComponentEvidence struct {
	label       string // label for the component evidence
	data        []byte // component evidence payload
	mediaType   string // media type of the component evidence
	parentLabel string // label of the parent component evidence (empty for root)
	depth       uint   // depth in the component evidence tree (0 for root)
}
