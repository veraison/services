// Copyright 2021-2023 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package tpm_enacttrust

import (
	"encoding/json"
	"fmt"
)

type TaAttr struct {
	NodeID string `json:"enacttrust-tpm.node-id"`
	Key    string `json:"enacttrust.ak-pub"`
}

type TrustAnchorEndorsement struct {
	Scheme  string `json:"scheme"`
	Type    string `json:"type"`
	SubType string `json:"sub_type"`
	Attr    TaAttr `json:"attributes"`
}

type RefValAttr struct {
	NodeID string `json:"enacttrust-tpm.node-id"`
	Digest string `json:"enacttrust-tpm.digest"`
	AlgId  int    `json:"enacttrust-tpm.alg-id"`
}

type RefValEndorsement struct {
	Scheme  string     `json:"scheme"`
	Type    string     `json:"type"`
	SubType string     `json:"sub_type"`
	Attr    RefValAttr `json:"attributes"`
}

type Endorsements struct {
	Digest string
}

func (e *Endorsements) Populate(strings []string) error {
	l := len(strings)

	if l != 1 {
		return fmt.Errorf("incorrect endorsements number: want 1, got %d", l)
	}

	var refval RefValEndorsement

	if err := json.Unmarshal([]byte(strings[0]), &refval); err != nil {
		return fmt.Errorf("could not decode reference value: %w", err)
	}

	e.Digest = refval.Attr.Digest

	return nil
}
