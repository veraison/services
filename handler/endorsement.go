// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package handler

import "encoding/json"

const (
	EndorsementType_UNSPECIFIED      string = "unspecified"
	EndorsementType_REFERENCE_VALUE  string = "reference value"
	EndorsementType_VERIFICATION_KEY string = "trust anchor"
)

type Endorsement struct {
	Scheme string `json:"scheme"`
	Type   string `json:"type"`

	SubType    string          `json:"subType"`
	Attributes json.RawMessage `json:"attributes"`
}
type EndorsementHandlerResponse struct {
	ReferenceValues []Endorsement
	TrustAnchors    []Endorsement
	SignerInfo      map[string]string
}
