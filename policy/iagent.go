// Copyright 2022-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package policy

import (
	"context"

	"github.com/spf13/viper"
	"github.com/veraison/corim/comid"
	"github.com/veraison/ear"
	"github.com/veraison/services/vts/appraisal"
)

type IAgent interface {
	Init(v *viper.Viper) error
	GetBackendName() string
	Evaluate(ctx context.Context,
		sessionContext map[string]any,
		appraisalContext *appraisal.Context,
		policy *Policy,
		submod string,
		appraisal *ear.Appraisal,
		endorsements []*comid.ValueTriple,
	) (*ear.Appraisal, error)
	Validate(ctx context.Context, policyRules string) error
	Close()
}
