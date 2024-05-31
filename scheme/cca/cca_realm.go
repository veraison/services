// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cca

import (
	"github.com/veraison/ear"
	"github.com/veraison/services/handler"
)

type Cca_realm_attester struct {
}

func (Cca_realm_attester) PerformAppraisal(
	appraisal *ear.Appraisal,
	ev map[string]interface{},
	endorsements []handler.Endorsement) error {

	return nil
}
