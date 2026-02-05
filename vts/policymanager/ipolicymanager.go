// Copyright 2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policymanager

import (
	"context"

	"github.com/veraison/services/vts/appraisal"
)

type IPolicyManager interface {
	Evaluate(ctx context.Context, appraisal *appraisal.Appraisal) error
}
