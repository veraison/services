// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package compositeevidenceparser

import "github.com/pkg/errors"

type eat struct{}

func (o eat) Parse(evidence []byte) ([]ComponentEvidence, error) {
	return nil, errors.New("not implemented")
}
