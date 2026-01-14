// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package compositeevidenceparser

import "github.com/pkg/errors"

type cmw struct{}

func (o cmw) Parse(evidence []byte) ([]ComponentEvidence, error) {
	return nil, errors.New("not implemented")
}
