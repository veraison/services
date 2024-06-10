// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package arm_cca

import (
	"github.com/veraison/ear"
	"github.com/veraison/services/handler"
)

type ISubAttester interface {
	PerformAppraisal(*ear.Appraisal, map[string]interface{}, []handler.Endorsement) error
}
